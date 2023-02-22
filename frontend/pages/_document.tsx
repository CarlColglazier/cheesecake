import { Html, Head, Main, NextScript } from 'next/document'
import Link from 'next/link'

export default function Document() {
  return (
    <Html lang="en">
      <Head />
      <body className="bg-white text-gray-700">
        <div className="container mx-auto px-4">
          <nav className='px-6'>
            <h1 className="text-3xl font-bold underline"><Link href="/events">Cheesecake</Link></h1>
          </nav>
          <section className='container mx-auto px-6 p-10'>
            <Main />
          </section>
          <NextScript />
        </div>
      </body>
    </Html>
  )
}
