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

import React from 'react';
import {Snackbar, SnackbarCloseReason, Button, IconButton} from "@mui/material";
import {useState} from "react";
import {MediaResult} from "../shared/model";
import SearchBar from "../components/SearchBar";
import MediaResults from "../components/MediaResults";
import CloseIcon from '@mui/icons-material/Close';

const Search = () => {
    const [results, setResults] = useState<Array<MediaResult>>([]);
    const [message, setMessage] = useState<string>(null!);
    const [open, setOpen] = useState<boolean>(false);

    const handleClose = (
        _: React.SyntheticEvent | Event,
        reason?: SnackbarCloseReason,
    ) => {
        if (reason === 'clickaway') {
            return;
        }
        setOpen(false);
    };

    const action = (
        <React.Fragment>
            <Button color="secondary" size="small" onClick={handleClose}>
                Close
            </Button>
            <IconButton
                size="small"
                aria-label="close"
                color="inherit"
                onClick={handleClose}
            >
                <CloseIcon fontSize="small" />
            </IconButton>
        </React.Fragment>
    );

    return (
        <>
            <SearchBar setMessage={setMessage} setOpen={setOpen} setResults={setResults}/>
            <MediaResults results={results}/>
            <Snackbar
                anchorOrigin={{vertical: 'top', horizontal: 'center'}}
                open={open}
                autoHideDuration={6000}
                onClose={handleClose}
                message={message}
                action={action}
            />
        </>
    );
};

export default Search;
