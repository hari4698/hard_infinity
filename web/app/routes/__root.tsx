// app/routes/__root.tsx
import type { ReactNode } from "react";
import {
  Outlet,
  createRootRoute,
  HeadContent,
  Scripts,
  Link,
} from "@tanstack/react-router";

import appCss from "@/styles/app.css?url";

function NotFound() {
  return (
    <div className="p-4 text-center">
      <h1 className="text-2xl font-bold mb-4">404 - Page Not Found</h1>
      <p className="mb-4">
        Sorry, the page you are looking for does not exist.
      </p>
      <Link to="/" className="text-blue-600 hover:underline">
        Go back home
      </Link>
    </div>
  );
}

export const Route = createRootRoute({
  head: () => ({
    meta: [
      {
        charSet: "utf-8",
      },
      {
        name: "viewport",
        content: "width=device-width, initial-scale=1",
      },
      {
        title: "Hard Infinity",
      },
    ],
    links: [
      {
        rel: "stylesheet",
        href: appCss,
      },
    ],
  }),
  component: RootComponent,
  notFoundComponent: NotFound,
});

function RootComponent() {
  return (
    <RootDocument>
      <div className="min-h-screen">
        <header>
          <div className="mx-auto max-w-3xl h-full p-4">
            <div className="flex items-center justify-between">
              <Link to="/" >
                <h1 className="text-4xl transform transition-transform duration-300 hover:scale-110">HardInfinity!</h1>
              </Link>
              <nav>
                <ul className="flex space-x-4">
                  <li>
                    <Link to="/settings" className="text-blue-600 hover:underline">
                      Settings
                    </Link>
                  </li>
                </ul>
              </nav>
            </div>
          </div>
        </header>
        <main className="flex-1 overflow-x-hidden overflow-y-auto">
          <div className="mx-auto max-w-3xl h-full p-4">
            <Outlet />
          </div>
        </main>
      </div>
    </RootDocument>
  );
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html>
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}
