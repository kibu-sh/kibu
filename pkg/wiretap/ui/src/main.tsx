import '@radix-ui/themes/styles.css'
import './index.css'

import { Theme } from '@radix-ui/themes'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'

import App from './App.tsx'
import { loader } from './AppLoader.ts'

const router = createBrowserRouter([
  {
    path: '/__admin/ui/',
    loader: loader,
    element: <App />,
  },
])

ReactDOM.createRoot(document.getElementById('root')!).render(
  <Theme appearance="dark">
    <React.StrictMode>
      <RouterProvider router={router} />
    </React.StrictMode>
  </Theme>,
)
