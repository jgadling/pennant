Build protobuf from the root of this repo:
protoc -I proto proto/pennant.proto --go_out=plugins=grpc:proto
protoc -I proto --python_out=tools/python --plugin=protoc-gen-grpc=/usr/local/bin/grpc_python_plugin proto/pennant.proto


Also, right now all document values in the grpc api are strings. This allows us
to use the built-in protobuf map type to accept arbitrary fields and values,
without throwing out TOO much functionality:
 - most gates don't do numeric comparisons against values
 - Regexes can take the place of some types of numeric comparisons
 - We can look for ways to add better support for int/float types in the future
