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

import {Paper, Typography} from "@mui/material"
import {MediaResult} from "../shared/model"
import MediaRow from "./MediaRow"

const MediaResults = ({results}: { results: MediaResult[] }) => {
    if (results && results.length > 0) {
        const mappedResults = results.map((r) => (<MediaRow key={r.id} result={r}/>))
        return (
            <Paper sx={{p: 2, mb: 2}} elevation={5}>
                <Typography variant="h4" sx={{mb: 2}}>Results</Typography>
                {mappedResults}
            </Paper>
        )
    } else {
        return (<></>)
    }
}
export default MediaResults