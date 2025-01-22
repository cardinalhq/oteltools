#!/bin/sh
# Copyright 2024-2025 CardinalHQ, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# only rebuild if missing, empty, or out of date
if [ ! -f tokenizer.go ] || [ ! -s tokenizer.go ] || [ tokenizer.rl -nt tokenizer.go ] || [ runragel.sh -nt tokenizer.go ]; then
    set -x
    ragel -Z -F1 tokenizer.rl -o /dev/stdout | sed '1 s,^.*$,// GENERATED CODE.  DO NOT EDIT.,' > tokenizer.go
fi
