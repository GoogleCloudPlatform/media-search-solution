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

import {Grid2, Typography} from "@mui/material";
import React from "react";
import {CastMember} from "../shared/model";

const Cast = ({cast}: { cast: CastMember[] }) => {
    return (
        <React.Fragment>
            <Typography variant="h6">Cast</Typography>
            <Grid2 container spacing={2} sx={{border: '1px solid #666', p: 1}}>
                <Grid2 size={8} sx={{fontWeight: 900, textAlign: 'center'}}>Character Name</Grid2>
                <Grid2 size={4} sx={{fontWeight: 900, textAlign: 'center'}}>Actor/Actress</Grid2>
                {cast?.map((c) => (
                    <>
                        <Grid2 size={8}><Typography variant="body2"><span
                            style={{fontWeight: 800, marginRight: '3em'}}>{c.character_name}</span></Typography></Grid2>
                        <Grid2 size={4}><Typography variant="body2">{c.actor_name}</Typography></Grid2>
                    </>
                ))}
            </Grid2>
        </React.Fragment>
    );
}

export default Cast
