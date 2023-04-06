import { Key } from "react"
import React from "react";
import Link from 'next/link'
import path from 'path';
import { promises as fs } from 'fs';
import { GetStaticProps, GetStaticPaths, GetServerSideProps } from 'next'

type Props = {
  events: any[],
  keys: readonly string[]
}

function construct_keymap(events: any[]) {
  let keymap: any = {};
  events.forEach((e: any) => {
    if (keymap[e.key] === undefined) {
      keymap[e.key] = e;
    }
  });
  return keymap;
}

const EventIndexPage: React.FC<Props> = ({ events, keys }) => {
  const event_data = construct_keymap(events);
  return (
    <>
      <ul>
        {keys.map((e: string, i: Key | null | undefined) => (
          <li key={i} ><Link href={"/events/" + e}>{event_data[e]["name"]}</Link></li>
        ))}
      </ul>
    </>
  )
}


export const getStaticProps: GetStaticProps = async () => {
  const jsonDirectory = path.join(process.cwd(), '../files/api/');
  const fileContents = await fs.readFile(jsonDirectory + `events.json`, 'utf8')
  const data = JSON.parse(fileContents);

  const key_fileContents = await fs.readFile(jsonDirectory + `event_keys.json`, 'utf8')
  const key_data = JSON.parse(key_fileContents);
  return {
    props: {
      events: data,
      keys: key_data
    }
  }
}

export default EventIndexPage;

