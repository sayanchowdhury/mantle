// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/mantle/Godeps/_workspace/src/google.golang.org/api/googleapi"
	"github.com/coreos/mantle/Godeps/_workspace/src/google.golang.org/api/storage/v1"
)

var (
	UnknownScheme = errors.New("storage: URL missing gs:// scheme")
	UnknownBucket = errors.New("storage: URL missing bucket name")
)

type Bucket struct {
	service *storage.Service
	name    string
	prefix  string

	mu      sync.RWMutex
	objects map[string]*storage.Object

	// writeAlways enables overwriting of objects that appear up-to-date
	writeAlways bool
	// writeDryRun blocks any changes, merely logging them instead
	writeDryRun bool
}

func NewBucket(client *http.Client, bucketURL string) (*Bucket, error) {
	service, err := storage.New(client)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(bucketURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "gs" {
		return nil, UnknownScheme
	}
	if parsedURL.Host == "" {
		return nil, UnknownBucket
	}

	// unsure if this is behavior is what we want or not
	if parsedURL.Path != "" && !strings.HasSuffix(parsedURL.Path, "/") {
		parsedURL.Path += "/"
	}

	return &Bucket{
		service: service,
		name:    parsedURL.Host,
		prefix:  strings.TrimPrefix(parsedURL.Path, "/"),
		objects: make(map[string]*storage.Object),
	}, nil
}

func (b *Bucket) Name() string {
	return b.name
}

func (b *Bucket) Prefix() string {
	return b.prefix
}

func (b *Bucket) URL() *url.URL {
	return &url.URL{Scheme: "gs", Host: b.name, Path: b.prefix}
}

func (b *Bucket) WriteAlways(always bool) {
	b.writeAlways = always
}

func (b *Bucket) WriteDryRun(dryrun bool) {
	b.writeDryRun = dryrun
}

// TrimPrefix removes the Bucket's path prefix from an Object name.
func (b *Bucket) TrimPrefix(objName string) string {
	if !strings.HasPrefix(objName, b.prefix) {
		panic(fmt.Errorf("%q missing prefix %q", objName, b.prefix))
	}
	return objName[len(b.prefix):]
}

// AddPrefix joins the Bucket's path prefix with a relative Object name.
func (b *Bucket) AddPrefix(objName string) string {
	return b.prefix + objName
}

func (b *Bucket) Object(objName string) *storage.Object {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.objects[objName]
}

func (b *Bucket) Objects() []*storage.Object {
	b.mu.RLock()
	defer b.mu.RUnlock()
	objs := make([]*storage.Object, 0, len(b.objects))
	for _, obj := range b.objects {
		objs = append(objs, obj)
	}
	return objs
}

func (b *Bucket) Prefixes() []string {
	prefixmap := make(map[string]struct{})
	prefixlist := make([]string, 0)

	b.mu.RLock()
	defer b.mu.RUnlock()
	for path := range b.objects {
		for strings.HasPrefix(path, b.prefix) {
			i := strings.LastIndexByte(path, '/')
			if i < 0 {
				break
			}
			path = path[:i+1]
			if _, ok := prefixmap[path]; ok {
				break
			}
			prefixmap[path] = struct{}{}
			prefixlist = append(prefixlist, path)
		}
	}
	return prefixlist
}

func (b *Bucket) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.objects)
}

func (b *Bucket) addObject(obj *storage.Object) {
	if obj.Bucket != b.name {
		panic(fmt.Errorf("adding gs://%s/%s to bucket %s", obj.Bucket, obj.Name, b.name))
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.objects[obj.Name] = obj
}

func (b *Bucket) addObjects(objs *storage.Objects) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, obj := range objs.Items {
		if obj.Bucket != b.name {
			panic(fmt.Errorf("adding gs://%s/%s to bucket %s", obj.Bucket, obj.Name, b.name))
		}
		b.objects[obj.Name] = obj
	}
}

func (b *Bucket) delObject(objName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.objects, objName)
}

func (b *Bucket) mkURL(obj interface{}) *url.URL {
	switch v := obj.(type) {
	case string:
		u := b.URL()
		u.Path = v
		return u
	case *storage.Object:
		u := b.URL()
		u.Path = v.Name
		if v.Bucket != "" {
			u.Host = v.Bucket
		}
		return u
	case *url.URL:
		return v
	case nil:
		return b.URL()
	default:
		panic(fmt.Errorf("unknown type %T", obj))
	}
}

func (b *Bucket) apiErr(op string, obj interface{}, e error) error {
	if _, ok := e.(*googleapi.Error); ok {
		return &Error{Op: op, URL: b.mkURL(obj).String(), Err: e}
	}
	return e
}

func (b *Bucket) Fetch(ctx context.Context) error {
	req := b.service.Objects.List(b.name)
	if b.prefix != "" {
		req.Prefix(b.prefix)
	}

	add := func(objs *storage.Objects) error {
		b.addObjects(objs)
		plog.Infof("Found %d objects under %s", b.Len(), b.URL())
		return nil
	}

	startRequest()
	defer stopRequest()
	plog.Noticef("Fetching %s", b.URL())

	return b.apiErr("storage.objects.list", nil, req.Pages(nil, add))
}

func (b *Bucket) Upload(ctx context.Context, obj *storage.Object, media io.ReadSeeker) error {
	// Calculate the checksum to enable upload integrity checking.
	if obj.Crc32c == "" {
		obj = dupObj(obj) // avoid editing the original
		if err := crcSum(obj, media); err != nil {
			return err
		}
	}

	old := b.Object(obj.Name)
	if !b.writeAlways && crcEq(old, obj) {
		return nil // up to date!
	}
	if b.writeDryRun {
		plog.Noticef("Would write %s", b.mkURL(obj))
		return nil
	}

	req := b.service.Objects.Insert(b.name, obj)
	req.Context(ctx)
	req.Media(media)

	// Watch out for unexpected conflicting updates.
	if old != nil {
		req.IfGenerationMatch(old.Generation)
	}

	startRequest()
	defer stopRequest()
	plog.Noticef("Writing %s", b.mkURL(obj))

	inserted, err := req.Do()
	if err != nil {
		return b.apiErr("storage.objects.insert", obj, err)
	}

	b.addObject(inserted)
	return nil
}

func (b *Bucket) Copy(ctx context.Context, src *storage.Object, dstName string) error {
	if src.Bucket == "" {
		panic(fmt.Errorf("src.Bucket is blank: %#v", src))
	}

	old := b.Object(dstName)
	if !b.writeAlways && crcEq(old, src) {
		return nil // up to date!
	}

	// It does work to pass src directly to the Rewrite API call, the
	// name and bucket values don't really matter, they just cannot be
	// blank for whatever reason. We make a copy just to get consistent
	// results, e.g. always use the destination bucket's default ACL.
	dst := dupObj(src)
	dst.Name = dstName
	dst.Bucket = b.name

	if b.writeDryRun {
		plog.Noticef("Would copy %s to %s", b.mkURL(src), b.mkURL(dst))
		return nil
	}

	req := b.service.Objects.Rewrite(
		src.Bucket, src.Name, dst.Bucket, dst.Name, src)
	req.Context(ctx)

	// Watch out for unexpected conflicting updates.
	if old != nil {
		req.IfGenerationMatch(old.Generation)
	}
	if src.Generation != 0 {
		req.IfSourceGenerationMatch(src.Generation)
	}

	startRequest()
	defer stopRequest()
	plog.Noticef("Copying %s to %s", b.mkURL(src), b.mkURL(dst))

	for {
		resp, err := req.Do()
		if err != nil {
			return b.apiErr("storage.objects.rewrite", dst, err)
		}
		if resp.Done {
			b.addObject(resp.Resource)
			return nil
		}
		req.RewriteToken(resp.RewriteToken)
	}
}

func (b *Bucket) Delete(ctx context.Context, objName string) error {
	if b.writeDryRun {
		plog.Noticef("Would delete %s", b.mkURL(objName))
		return nil
	}

	req := b.service.Objects.Delete(b.name, objName)
	req.Context(ctx)

	// Watch out for unexpected conflicting updates.
	if old := b.Object(objName); old != nil {
		req.IfGenerationMatch(old.Generation)
		req.IfMetagenerationMatch(old.Metageneration)
	}

	startRequest()
	defer stopRequest()
	plog.Noticef("Deleting %s", b.mkURL(objName))

	if err := req.Do(); err != nil {
		return b.apiErr("storage.objects.delete", objName, err)
	}

	b.delObject(objName)
	return nil
}