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

export interface CastMember {
    actor_name: string;
    character_name: string;
}

export interface Segment {
    sequence: number;
    start: string;
    end: string;
    script: string;
}

export interface MediaResult {
    id: string;
    create_date: Date;
    title: string;
    category: string;
    summary: string;
    media_url: string;
    length_in_seconds: number;
    director: string;
    release_year: number;
    genre: string;
    rating: string;
    cast: CastMember[];
    segments: Segment[];
}
