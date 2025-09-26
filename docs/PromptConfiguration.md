# Media Search Prompt Configuration

## 1. Overview

This document explains how to configure and customize prompts in the Media Search Solution to tailor the analysis for your specific media content.

The media analysis pipeline uses three types of prompts, corresponding to the main analysis steps performed by the solution's AI models:

| Prompt Type          | Description                                                              |
| -------------------- | ------------------------------------------------------------------------ |
| Content Type Prompt  | Determines the content type of the media file (e.g., "sports", "trailer"). |
| Media Summary Prompt | Creates a content summary and identifies logical segments in the media file. |
| Segment Summary Prompt | Generates a detailed description for a specific segment. |



## 2. Configuration File

All prompts are configured in a TOML file located in the `configs/` directory. The base configuration is in `configs/.env.toml`. You can create environment-specific overrides, such as `configs/.env.local.toml` for local development.

The configuration logic is as follows:
1.  You define a list of supported content types (e.g., "sports", "trailer").
2.  You provide a prompt template that the AI uses to classify a video into one of those types.
3.  For each content type, you provide a set of specific prompt templates for generating summaries and segment descriptions.

The solution comes with two out-of-the-box content types:
- trailer (default type)
- sports

## 3. Prompt Structure
The configuration file uses TOML syntax. Below is a breakdown of the structure with examples.

### 3.1. Content Type Definition

The `[content_type]` table defines how the solution identifies the type of video.

```toml
[content_type]
# A list of all supported content types.
types = ["trailer", "sports"]

# The default type to use if detection fails. Must be in the `types` list.
default_type = "trailer"

# The prompt template used by the AI to determine the content type.
prompt_template = """
Analyze the content of the provided media file and determine its primary content type from the list below.
Your response must be one of the provided values, with an exact match. For the task it is enough to only analyze the first 30 seconds of the media file.

Available content types: {{ .CONTENT_TYPES }}

Do not add any other text, explanation, or formatting to your response. Only output the determined content type."""
```

**Template Variables for `prompt_template`:**

*   `{{ .CONTENT_TYPES }}`: Injected with the list of content types from the `types` array.

### 3.2. Prompt Templates per Content Type

For each content type defined in the `types` array, you must create a corresponding `[prompt_templates.{content_type}]` table. This table contains the specific instructions and prompts for that type.

```toml
# Example for the "trailer" content type.
[prompt_templates.trailer]
system_instructions = """
Your role is a film, and media trailer official capable of describing
in detail directors, producers, cinematographers, screenwriters, and actors.
In addition, you're able to summarize plot points, identify scene time stamps
and recognize which actor is playing which character, and which character is in each scene.
"""

summary = """Review the attached media file and extract the following information
- Title as title
- Lower case category name as category from one of the following categories and definitions:
    - {{ .CATEGORIES }}
- Summary - a detailed summary of the media contents, plot, and cinematic themes in markdown format
- Length in Seconds as length_in_seconds,
- Media URL as media_url
- Director as director
- Release Year as release_year, a four digit year
- Genre as genre
- Rating as rating with one of the following values: G, PG, PG-13, R, NC-17
- Cast as cast, an array of Cast Members including Character Name as character_name, and associated actor name as actor_name
- Extract the scenes based on their narrative and visual coherence, ordering them by start and end times. The primary goal is to create segments that feel natural and complete.
    - A scene is a continuous segment of action or dialogue in a single location. A scene break MUST occur at a logical transition point, such as:
        - A change in location.
        - A significant jump forward or backward in time.
        - The start/end of a major conversation or action sequence.
    - Crucially, DO NOT end a scene abruptly. Avoid cutting in the middle of a continuous camera shot (a single take) or in the middle of a spoken sentence.
    While scenes must have a minimum length of 10 seconds, their duration should be determined by the content. Prioritize logical, coherent segmentation over adhering to any specific length.
    - The segmented scenes must be continuous and cover the entire video from start to finish without any gaps. The total length of the video is {{ .VIDEO_LENGTH }} seconds.
    - The first scene must start at 00:00:00.
    - The end of one scene must be the exact start of the next scene.
    - The end time of the final scene must be the total duration of the video.
    - Add a sequence number to each scene starting from 1 and incrementing in order of the timestamp.

**Timestamp Formatting and Logic Rules:**
- All `start` and `end` timestamps must be strings formatted as "HH:MM:SS", with each component zero-padded to two digits. Values must be calculated correctly; for example, a moment 119 seconds into a video is "00:01:59", not "01:19:00".
- All timestamps must be logical and fall within the video's total duration. A video that is 1 minute and 59 seconds long cannot have a timestamp of "00:02:00" or greater.
- For any given scene, the `end` timestamp must always be chronologically after its `start` timestamp.

Example Output Format:
{{ .EXAMPLE_JSON }}
"""

segment = """Given the following media file, summary, actors, and characters, extract the following details for the time segment {{ .TIME_START }} - {{ .TIME_END }} in a valid JSON format.
The given time segement timestamps are in the format of HH:MM:SS or hours:minutes:seconds.
**Extraction Details:**
- sequence_number: {{ .SEQUENCE }} as a number
- start: {{ .TIME_START }} as a string
- end: {{ .TIME_END }} as a string
- script: write a detailed scene description that includes colors, action sequences, dialogue with both character and actor citations, any products or brand names, and lastly any significant props, in plain text.

**IMPORTANT FALLBACK INSTRUCTION:**
If you are unable to generate a detailed scene description from the video segment (for example, if the segment is too short, lacks distinct action, or has no dialogue), you MUST provide a default scene extraction. For this default scene, use the provided 'Media Summary' as the content for the 'script' field. The 'start' and 'end' times should still match the provided time frame {{ .TIME_START }} - {{ .TIME_END }}.

Media Summary:
{{ .SUMMARY_DOCUMENT }}

Example Output:
{{ .EXAMPLE_JSON }}"""
```

**Fields:**

*   `system_instructions`: Provides a role or context for the AI model when it processes the media file for this content type.
*   `summary`: The prompt template for generating a structured summary of the entire video.
*   `segment`: The prompt template for generating a structured description of a specific video segment.

**Template Variables for `summary` prompt:**

*   `{{ .VIDEO_LENGTH }}`: The total length of the video in seconds.
*   `{{ .CATEGORIES }}`: An optional field. If specified, a list of predefined categories and their definitions will be injected here.
*   `{{ .EXAMPLE_JSON }}`: An example JSON object to specify the expected output format.


**Template Variables for `segment` prompt:**

*   `{{ .SEQUENCE }}`: The sequence number of the segment being analyzed.
*   `{{ .TIME_START }}`: The start time of the segment segment in `HH:MM:SS` format.
*   `{{ .TIME_END }}`: The end time of the segment segment in `HH:MM:SS` format.
*   `{{ .SUMMARY_DOCUMENT }}`: The full media summary generated in the previous step.
*   `{{ .EXAMPLE_JSON }}`: An example JSON object to specify the expected output format.

### 3.3. JSON Output Schema
The JSON schemas for both the summary and segment outputs are defined in `pkg/model/schemas.go`. This file acts as the source of truth for the expected JSON structure. When you modify or create prompts, ensure that the fields you ask the AI to extract align with the definitions in the schema file to ensure correct parsing.


## 4. Customizing Prompts

You can easily add support for new content types or modify existing ones.

### 4.1. Modifying Existing Prompts

To change the analysis behavior for an existing content type (e.g., `sports`), edit the `system_instructions`, `summary`, or `segment` values within the corresponding `[prompt_templates.sports]` table in your configuration file.

### 4.2. Adding a New Content Type

To add a new content type, such as "news_report", follow these steps:

**Step 1: Update the Content Type List**

In the `[content_type]` table, add your new type to the `types` array.

```toml
[content_type]
types = ["trailer", "sports", "news_report"]
# ... rest of the table
```

**Step 2: Create a New Prompt Template Table**

Add a new table to your configuration file for your custom type.

```toml
[prompt_templates.news_report]
system_instructions = "You are a news editor. Your task is to summarize news reports and identify key segments."
summary = """
Analyze the attached media file containing a news report and extract the following information:
...

The video is {{ .VIDEO_LENGTH }} seconds long, etc
...

"""
segment = """
Given the following media file, analyze and describe the news segment from {{ .TIME_START }} to {{ .TIME_END }}.
**Extraction Details:**
...
Media Summary:
{{ .SUMMARY_DOCUMENT }}

Example Output:
{{ .EXAMPLE_JSON }}"""

```

## Upload Configuration Changes

To apply prompt customizations, upload the modified `configs/.env.toml` file to the configuration bucket in Cloud Storage using the following command from the project root:

```sh
gsutil cp configs/.env.toml gs://$(terraform -chdir="build/terraform" output -raw config_bucket)/.env.toml
```

The Media Search service automatically detects this change, reloads the configuration, and uses the new prompts for all subsequent video processing.
