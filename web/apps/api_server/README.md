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

# API Server

This is a simple server housing multiple functions

* /media?s= search
* /media/:id find media by id
* /media/:id/segments/:segment_id find segments

## Prior to running the server

Make sure you create a local config file in "//configs/.env.local.toml".
In that configuration add the following filling in your project id

```toml
[application]
google_project_id=""

[big_query_data_sources."media_ds"]
dataset="media_ds"
table_names = ["media"]

[topic_subscriptions."HiResTopic"]
name="media_high_res_resources_subscription"
dead_letter_topic="media_high_res_events_dead_letter"
timeout_in_seconds=10
command_name=""

[topic_subscriptions."LowResTopic"]
name="media_low_res_resources_subscription"
dead_letter_topic="media_low_res_events_dead_letter"
timeout_in_seconds=10
command_name=""
```

## Running the server

```shell

# Start the server on port 8080
bazel run //web/apps/api_server

```

## Example output

```json
[
  {
    "seq": 14,
    "start": "00:00:58",
    "end": "00:01:02",
    "script": "EXT. DESERT ROAD - DAY\n\nFour figures in black tactical gear are running across a dry riverbed.\n\nVOICEOVER (V.O.) - (Tom Hardy)\nThis is major...\n\nThe camera pans up to show them running across a rocky hillside.\n\nVOICEOVER (V.O.) - (Tom Hardy)\n...we are...\n\nOne of the figures dives into a body of water. Another figure dives in after him."
  },
  {
    "seq": 12,
    "start": "00:00:53",
    "end": "00:00:56",
    "script": "INT. BAR - NIGHT\n\nAnne Weying (Michelle Williams) is talking to a man.\n\nANNE WEYING - (Michelle Williams)\nLet’s go get them.\n\nEXT. STREET - NIGHT\nA group of soldiers are walking down a street at night. They are carrying weapons and flashlights. They look determined.\n\nSOLDIER 1 - (unidentified)\nOh, shit!"
  },
  {
    "seq": 11,
    "start": "00:00:50",
    "end": "00:00:53",
    "script": "INT. LABORATORY - DAY\n\nAnne Weying (V.O.) - (Michelle Williams)\nAnd it's our job to make sure that remains a secret.\n\nWe see a close up of Anne Weying, looking concerned.\n\nINT. BAR - NIGHT\n\nTwo men sit at a bar, looking serious. One of them is a military man.\n\nINT. LABORATORY - DAY\n\nAnne Weying is talking to a military man. She looks worried."
  },
  {
    "seq": 5,
    "start": "00:00:17",
    "end": "00:00:20",
    "script": "EXT. WAREHOUSE - NIGHT\n\nEDDIE BROCK (V.O.) - (Tom Hardy)\nWhat?\n\nVenom is shown emerging from Eddie Brock's body, his face partially visible.\n\nVENOM - (Tom Hardy)\nWe are Venom!\n\nFour men are seen approaching, carrying weapons. They look determined and ready for a fight."
  },
  {
    "seq": 19,
    "start": "00:01:12",
    "end": "00:01:16",
    "script": "INT. WAREHOUSE - NIGHT\n\nEddie Brock is held at gunpoint by several armed men.\n\nMAN 1 - (Unidentified)\nSay when.\n\nEddie looks at the men, a knife appears in his hand, and he attacks.\n\nEDDIE BROCK - (Tom Hardy)\nWhen."
  }
]

```

