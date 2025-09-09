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

import {FormControl, IconButton, InputAdornment, InputLabel, OutlinedInput, Paper,} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import {MediaResult} from "../shared/model";
import axios from "axios";
import {useState} from "react";

type SearchBarArgs = {
    setResults: (values: MediaResult[]) => void
    setMessage: (value: string) => void
    setOpen: (value: boolean) => void
}

const SearchBar = ({setResults, setMessage, setOpen}: SearchBarArgs) => {
    const runQuery = () => {
        setMessage("Searching...")
        setOpen(true)
        setResults([])
        const baseURL = process.env.NODE_ENV === "development" ? "http://localhost:8080" : "";

        axios
            .get(`${baseURL}/api/v1/media?count=5&s=${query}`)
            .then((r) => {
                console.log(r)
                if (r.status == 200) {
                    setResults([...r.data]);
                } else {
                    setMessage(
                        `Invalid HTTP Response: ${r.status} ${r.statusText} - ${r.data}`,
                    );
                    setOpen(true);
                }
            })
            .catch((e) => {
                setMessage(e);
                setOpen(true);
            });
    };

    const [query, setQuery] = useState<string>(null!);

    const keyUp = (e: React.KeyboardEvent<HTMLTextAreaElement | HTMLInputElement>) => {
        if (e.key === 'Enter') {
            runQuery()
        }
    }

    return (
        <Paper
            sx={{p: 2, mb: 2, display: "flex", flexDirection: "row", flex: 1}}
            elevation={4}
        >
            <FormControl variant="outlined" fullWidth>
                <InputLabel htmlFor="search-adornment">Search</InputLabel>
                <OutlinedInput
                    id="search-adornment"
                    type="text"
                    onChange={(v) => setQuery(v.target.value)}
                    onKeyDown={keyUp}
                    endAdornment={
                        <InputAdornment position="end">
                            <IconButton
                                sx={{p: "10px", mt: "2px"}}
                                aria-label="search"
                                onClick={runQuery}
                            >
                                <SearchIcon/>
                            </IconButton>
                        </InputAdornment>
                    }
                    label="Search"
                />
            </FormControl>
        </Paper>
    );
};

export default SearchBar;
