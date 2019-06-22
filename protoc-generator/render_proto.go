// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protoc_generator

import (
	"github.com/golang/protobuf/descriptor"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/ptypes/empty"
	prDesc "github.com/jhump/protoreflect/desc"
	prPrint "github.com/jhump/protoreflect/desc/protoprint"
	"google.golang.org/genproto/googleapis/api/annotations"
)

func (renderer *Renderer) RenderProto(fdSet *dpb.FileDescriptorSet) ([]byte, error) {

	buildDependenciesForProtoReflect(fdSet)

	// Creates a protoreflect FileDescriptor, which is then used for printing.
	prFd, err := prDesc.CreateFileDescriptorFromSet(fdSet)
	if err != nil {
		return nil, err
	}

	// Print the protoreflect FileDescriptor.
	p := prPrint.Printer{}
	res, err := p.PrintProtoToString(prFd)
	if err != nil {
		return nil, err
	}

	f := NewLineWriter()
	f.WriteLine(res)

	return f.Bytes(), err
}

// Protoreflect needs all the dependencies that are used inside of the initial FileDescriptorProto
// to work properly. Those dependencies are google/protobuf/empty.proto, google/api/annotations.proto,
// and "google/protobuf/descriptor.proto". For all those dependencies the corresponding
// FileDescriptorProto has to be added to the FileDescriptorSet. Protoreflect won't work
// if a reference is missing.
func buildDependenciesForProtoReflect(fdSet *dpb.FileDescriptorSet) {
	// Dependency to "google/protobuf/empty.proto" for RPC methods without any request / response
	// parameters.
	e := empty.Empty{}
	fd, _ := descriptor.ForMessage(&e)

	// Dependency to google/api/annotations.proto. Here a couple of problems arise:
	// 1. Problem: 	The name is set wrong
	// 2. Problem: 	We cannot call descriptor.ForMessage(&annotations.E_Http), which would be our
	//				required dependency. However, we can call descriptor.ForMessage(&http) and
	//				then construct the extension manually.
	// 3. Problem: 	.google.protobuf.MethodOptions gets extended, which is described inside
	//				"google/protobuf/descriptor.proto", therefore we need to add it as dependency.
	http := annotations.Http{}
	fd2, _ := descriptor.ForMessage(&http)

	extensionName := "http"
	n := "google/api/annotations.proto"
	l := dpb.FieldDescriptorProto_LABEL_OPTIONAL
	t := dpb.FieldDescriptorProto_TYPE_MESSAGE
	tName := "google.api.HttpRule"
	extendee := ".google.protobuf.MethodOptions"

	httpExtension := &dpb.FieldDescriptorProto{
		Name:     &extensionName,
		Number:   &annotations.E_Http.Field,
		Label:    &l,
		Type:     &t,
		TypeName: &tName,
		Extendee: &extendee,
	}

	fd2.Name = &n                                                               // 1. Problem
	fd2.Extension = append(fd2.Extension, httpExtension)                        // 2. Problem
	fd2.Dependency = append(fd2.Dependency, "google/protobuf/descriptor.proto") //3.rd Problem

	// Dependency to google/protobuf/descriptor.proto to address 3.rd Problem. FileDescriptorProto
	// still needs to be added.
	fdp := dpb.FieldDescriptorProto{}
	fd3, _ := descriptor.ForMessage(&fdp)

	// According to the documentation of prDesc.CreateFileDescriptorFromSet the file I want to print
	// needs to be at the end of the array. All other FileDescriptorProto are dependencies.
	fdSet.File = append([]*dpb.FileDescriptorProto{fd, fd2, fd3}, fdSet.File...)

}

//TODO: handle enum. Not sure if possible, because of
//TODO: https://github.com/googleapis/googleapis/blob/a8ee1416f4c588f2ab92da72e7c1f588c784d3e6/google/api/http.proto#L62

//TODO: Removing duplicates from responses
//TODO: Regenerate mode for proto testing
