<!--
 Copyright 2024 Google, LLC
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 
     https://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Test

The test directory contains all test cases for the media workflows, if you'd like to test manually,
and have a clean environment execute the following in order:

```shell
bazel build //...

# Reads the test media file
bazel test //test/workflow:media_ingestion

# Create the embeddings
bazel test //test/workflow:media_embeddings

# Run a test cast, find all segments with Woody Harrelson AKA Carnage
bazel test //test/services/...

```

## Example Output

Here you should see the media file ID, and the segment sequence number.

```shell
0192aa4e-e375-7f26-aaec-ebe4fb4d803b - 25
0192aa4e-e375-7f26-aaec-ebe4fb4d803b - 13
0192aa4e-e375-7f26-aaec-ebe4fb4d803b - 12
0192aa4e-e375-7f26-aaec-ebe4fb4d803b - 19
0192aa4e-e375-7f26-aaec-ebe4fb4d803b - 28
```
