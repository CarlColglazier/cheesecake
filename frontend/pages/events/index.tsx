import { Key } from "react"
import React from "react";
import Link from 'next/link'
import path from 'path';
import { promises as fs } from 'fs';
import { GetStaticProps, GetStaticPaths, GetServerSideProps } from 'next'

type Props = {
  events: readonly string[]
}

const EventIndexPage: React.FC<Props> = ({ events }) => {
  return (
    <>
      <p>This is the page for the event index.</p>
      <ul>
        {events.map((e: string, i: Key | null | undefined) => (
          <li key={i} ><Link href={"/events/" + e}>{e}</Link></li>
        ))}
      </ul>
    </>
  )
}


export const getStaticProps: GetStaticProps = async () => {
  const jsonDirectory = path.join(process.cwd(), '../files/api/');
  const fileContents = await fs.readFile(jsonDirectory + `events.json`, 'utf8')
  const data = JSON.parse(fileContents);
  return {
    props: {
      events: data
    }
  }
}

export default EventIndexPage;

