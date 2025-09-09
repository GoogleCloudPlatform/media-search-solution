/**
 * Copyright 2025 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Author: rrmcguinness (Ryan McGuinness)
 */

import "./App.css";
import {Box, Container, createTheme, CssBaseline, ThemeOptions, ThemeProvider, Typography} from "@mui/material";
import {LoadingIcon} from "./components/LoadingIcon";
import React from "react";
import {Outlet} from "react-router-dom";
import Footer from "./components/Footer";
import TopNav from "./components/TopNav";

const themeOptions: ThemeOptions = {
    palette: {
        mode: 'dark',
        primary: {
            main: '#1565c0',
        },
        secondary: {
            main: '#37474f',
        },
        error: {
            main: '#b71c1c',
        },
        warning: {
            main: '#f4511e',
        },
    },
};

const theme = createTheme(themeOptions);

function App() {
    return (
        <ThemeProvider theme={theme}>
            <TopNav/>
            <Typography variant="h3" sx={{ml: 2, mt: 2, fontFamily: 'Google Sans', fontWeight: 800, color: '#4285F4'}}>Me<span
                style={{color: '#FBBC04'}}>d</span>ia <span style={{color: '#DB4437'}}>S</span>ea<span
                style={{color: '#0F9D58'}}>r</span>ch</Typography>
            <Box
                sx={{
                    position: "relative",
                    display: "flex",
                    flexDirection: "column",
                    minHeight: "100vh",
                    minWidth: "100vw"

                }}
            >
                <CssBaseline/>
                <Container
                    component="main"
                    sx={{mt: 3, pb: "3.5em", mb: 2}}
                    maxWidth="xl"
                >
                    <React.Suspense fallback={<LoadingIcon/>}>
                        <Outlet/>
                    </React.Suspense>
                </Container>
                <Footer/>
            </Box>
        </ThemeProvider>
    );
}

export default App;
