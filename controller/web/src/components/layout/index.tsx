"use client";

import { JSXElementConstructor, ReactElement, ReactNode } from "react";

import Sidebar from "../sidebar";

type LayoutProps = Required<{
  readonly children: ReactNode;
}>;

export default function Layout({ children }: LayoutProps) {
  return (
    <>
      <Sidebar>
        <main>{children}</main>
      </Sidebar>
    </>
  );
}
