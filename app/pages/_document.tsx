import { Html, Head, Main, NextScript } from 'next/document'
import * as React from 'react';

export default function Document() {
  return (
    <Html lang="en" className={'bg-black text-white h-screen'}>
      <Head />
        <body>
            <Main />
            <NextScript />
        </body>
    </Html>
  )
}
