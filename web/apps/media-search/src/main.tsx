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

import React from 'react';
import {createRoot} from 'react-dom/client'
import App from './App.tsx'
import './index.css'

import {createBrowserRouter, Navigate, RouterProvider} from 'react-router-dom';

const Search = React.lazy(() => import('./pages/Search'));
const FileUpload = React.lazy(() => import('./pages/FileUpload'));
const Dashboard = React.lazy(() => import('./pages/Dashboard'))

const ErrorBoundary = () => {
    return (<Navigate to="/"/>)
}

const router = createBrowserRouter([
    {
        path: "/",
        element: <App/>,
        children: [
            {
                index: true,
                element: <Search/>,
                errorElement: <ErrorBoundary/>,
            },
            {
                path: "/uploads",
                element: <FileUpload/>,
                errorElement: <ErrorBoundary/>,
            },
            {
                path: "/dashboard",
                element: <Dashboard/>,
                errorElement: <ErrorBoundary/>,
            }
        ]
    }
])

createRoot(document.getElementById('root')!).render(<RouterProvider router={router}/>)