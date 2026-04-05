import "./globals.css";

export const metadata = {
  title: "多Agent智能合同审查系统",
  description: "基于LangGraph的四Agent流水线合同审查系统",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
