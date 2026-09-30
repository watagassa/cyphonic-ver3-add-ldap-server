import { Metadata } from "next";
import { Providers } from "@/app/providers";
import Layout from "@/components/layout";

export const metadata: Metadata = {
  title: "CYPHONIC Cloud Controller",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body>
        <Providers>
          <Layout>{children}</Layout>
        </Providers>
      </body>
    </html>
  );
}
