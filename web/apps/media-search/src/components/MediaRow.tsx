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

import {Box, Grid2, Typography} from "@mui/material";
import SegmentData from "./SegmentData";
import {MediaResult, Segment} from "../shared/model";
import Cast from "./Cast";

const MediaRow = ({result}: { result: MediaResult }) => {
    return (
        <Grid2 container spacing={1} sx={{pb: 4}}>
            <Grid2 size={4} sx={{textAlign: 'left', padding: 1}}>
                <Typography variant="h5" sx={{mb: 1}} color={"info"}>{result.title}</Typography>
                <Typography variant="h6">Summary</Typography>
                <Box sx={{pl: 2, pr: 2}}>
                    <Typography variant="caption">{result.summary}</Typography>
                </Box>
                <Cast cast={result.cast}/>
            </Grid2>
            <Grid2 size={8}>
                {result.segments.map((s: Segment, j:number) => (
                    <Grid2 container spacing={2} sx={{p: 1, mb: 3}} key={`result_${result.id}_${j}`}>
                        <SegmentData key={`${result.id}-${s.sequence}`} url={result.media_url}  segment={s}/>
                    </Grid2>
                ))}
            </Grid2>
        </Grid2>
    );
};

export default MediaRow