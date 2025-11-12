// Copyright 2025 Google, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: rrmcguinness (Ryan McGuinness)
//         kingman (Charlie Wang)

package main

import (
	"log"

	"github.com/GoogleCloudPlatform/media-search-solution/analyze/common"
)

func main() {
	genaiRunConfig, err := common.NewGenaiRunConfig()
	if err != nil {
		log.Fatal(err)
	}
	get_content_length(&genaiRunConfig.BasicRunConfig)
	get_content_type(genaiRunConfig)
	get_content_summary(genaiRunConfig)
	get_segment_summaries(genaiRunConfig)
	persist_analysis_result(genaiRunConfig)
	generate_embeddings(genaiRunConfig)

}
