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

import React from "react";
import {
    Box,
    Button,
    Container,
    Grid2,
    IconButton,
    List,
    ListItem,
    Snackbar,
    SnackbarCloseReason,
    Stack,
    Typography
} from "@mui/material";
import {useState} from "react";
import {FileUploader} from "react-drag-drop-files";
import "./FileUpload.css"
import axios from "axios";
import CloseIcon from '@mui/icons-material/Close';

const SupportedFileTypes = ["mp4"]

const FileListView = ({listOfFiles}: {listOfFiles: FileList}) => {
    if (listOfFiles != null) {
        const out = []
        for (let i = 0; i < listOfFiles.length; i++) {
            out.push(<ListItem>{listOfFiles[i].name}</ListItem>)
        }

        return (
            <List sx={{pt: 0, mt: 0}}>
                {out.map(item => item)}
            </List>
        )
    }
    return(<Container>No Videos</Container>)
}

const FileUpload = () => {
    const [open, setOpen] = useState(false);
    const [files, setFiles] = useState<FileList>(null!);

    const handleChange = (file: any) => {
        setFiles(file);
    }
    const onDrop = (file: any) => {
        console.log('drop', file);
    };
    const onSelect = (file: any) => {
        console.log('test', file);
    };

    const handleClose = (
        _: React.SyntheticEvent | Event,
        reason?: SnackbarCloseReason,
    ) => {
        if (reason === 'clickaway') {
            return;
        }
        setOpen(false);
    };

    const onTypeError = (err = 1) => console.log(err);
    const onSizeError = (err = 1) => console.log(err);

    const submitData = () => {
        if (!files) {
            return
        }
        const form = new FormData()
        for (const file of files) {
            form.append("files", file)
        }
        const baseURL = process.env.NODE_ENV === "development" ? "http://localhost:8080" : "";
        axios.post(`${baseURL}/api/v1/uploads`, form, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        }).then(r => {
            setOpen(true);
            setFiles(null!)
            console.log(r)
        }).catch(e => {
            console.log(e)
        })
    }

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
        <Container>
            <Stack spacing={2}>
                <Typography variant={'h5'}>Upload video file(s)</Typography>
                <Grid2 container>
                    <Grid2 size={6}>
                        <FileUploader
                            classes="upload-files"
                            fileOrFiles={files}
                            onTypeError={onTypeError}
                            handleChange={handleChange}
                            name="file"
                            types={SupportedFileTypes}
                            onSizeError={onSizeError}
                            onDrop={onDrop}
                            onSelect={onSelect}
                            label="Upload file here"
                            dropMessageStyle={{backgroundColor: '#34A853'}}
                            multiple
                        />
                    </Grid2>
                    <Grid2 size={6} sx={{verticalAlign: 'top', mt: 0, pt: 0}}>
                        <FileListView listOfFiles={files} />
                    </Grid2>
                </Grid2>

                <Box sx={{mt: 2, display: 'flex', flex: 1, justifyContent: 'right'}}>
                <Button variant="contained" sx={{width: '50%'}} onClick={submitData}>Submit</Button>
                </Box>
            </Stack>

            <Snackbar
                anchorOrigin={{vertical: 'top', horizontal: 'center'}}
                open={open}
                autoHideDuration={6000}
                onClose={handleClose}
                message="Files uploaded, it will take a few minutes to process."
                action={action}
            />
        </Container>
    )
}

export default FileUpload;