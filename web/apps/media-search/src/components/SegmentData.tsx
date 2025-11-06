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
import {Segment} from "../shared/model";

const SegmentData = ({url, segment}: { url: string, segment: Segment }) => {

    const formatScript = (val: string): string => {
        return val.replace("\n", "<br/>")
    }

    const GetStartTimeInSeconds = (): number => {
        const parts = segment.start.split(':');
        return parseInt(parts[0])*60*60 + parseInt(parts[1])*60 + parseInt(parts[2]);
    }

    const GetEndTimeInSeconds = (): number => {
        const parts = segment.end.split(':');
        return parseInt(parts[0])*60*60 + parseInt(parts[1])*60 + parseInt(parts[2]);
    }

    return (
        <>
            <Grid2 size={6}>
                <Grid2 container spacing={2}>
                    <Grid2 size={4} sx={{fontWeight: 800}}>Sequence</Grid2>
                    <Grid2 size={4} sx={{fontWeight: 800}}>Start</Grid2>
                    <Grid2 size={4} sx={{fontWeight: 800}}>End</Grid2>

                    <Grid2 size={4}>{segment.sequence}</Grid2>
                    <Grid2 size={4}>{segment.start}</Grid2>
                    <Grid2 size={4}>{segment.end}</Grid2>
                </Grid2>
            </Grid2>
            <Grid2 size={6} >
                <Box sx={{display: 'flex', flex: 1, flexGrow: 1, justifyContent: 'center', justifyItems: 'center', alignItems: 'center', alignContent: 'center', padding: 2}}>
                <video controls style={{border: '1px solid #4285F4  ', borderRadius: '10px', boxShadow: '1px 1px 6px 1px #666'}}>
                    <source src={`${url}#t=${GetStartTimeInSeconds()},${GetEndTimeInSeconds()}`} type="video/mp4" />
                </video>
                </Box>
            </Grid2>
            <Grid2 size={12} sx={{textAlign: 'left'}}><Typography component="div" variant="body2">
                <div dangerouslySetInnerHTML={{__html: formatScript(segment.script)}}/>
            </Typography></Grid2>
        </>)
};

export default SegmentData