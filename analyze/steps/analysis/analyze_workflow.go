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
