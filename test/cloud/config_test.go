// Copyright 2024 Google, LLC
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

package cloud_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/media-search-solution/test"
	"github.com/stretchr/testify/assert"
)

// TestConfig is used to test the validity of the hierarchy loader.
// First load is .env.toml, then .env.test.toml (set in test.SetupOS)
// any value redefined in .env.test.toml will overwrite .env.toml allowing
// the environment to take precedence over the defaults.
func TestConfig(t *testing.T) {
	config := test.GetConfig()
	// Uncomment this to see the final configuration structure
	// cloud.PrintConfig(config)

	assert.NotNil(t, config)
	assert.Equal(t, 2, len(config.TopicSubscriptions))
	assert.Equal(t, 2, len(config.EmbeddingModels))
	assert.Equal(t, 4, len(config.AgentModels))
	assert.Equal(t, 5, len(config.Categories))
}
