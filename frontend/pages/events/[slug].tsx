import { GetStaticProps, GetStaticPaths, GetServerSideProps } from 'next'
import path from 'path';
import { promises as fs } from 'fs';
import EVPlot from '../../components/EVPlot';
import MatchTable from '../../components/MatchTable';
import {EventDataType} from '../../types';
import { Tab } from '@headlessui/react';
import EVTable from '../../components/EVTable';
import useSWR, { SWRConfig, Fetcher } from 'swr';


/*
const fetcher: Fetcher<EventDataType, string> = (id) => {
  return fetch(`/api/events/${id}`)
    .then((res) => res.json)
}
*/

export const getStaticProps : GetStaticProps = async({ params }) => {
  const jsonDirectory = path.join(process.cwd(), '../files/api/events/');
  if (params !== undefined) {
    const fileContents = await fs.readFile(jsonDirectory + `${params.slug}.json`, 'utf8')
    const data = JSON.parse(fileContents);
    return {
      props: {
        slug: params.slug,
        fallback: data
      }
    }
  } else {
    return { 
      props: {
        slug: undefined
      } 
    }
  }
}

export async function getStaticPaths() {
  const jsonDirectory = path.join(process.cwd(), '../files/api/');
  const jsonEventsDirectory = path.join(process.cwd(), '../files/api/events/');
  const fileContents = await fs.readFile(jsonDirectory + '/events.json', 'utf8')
  const events = JSON.parse(fileContents)

  const paths = events.map((e: string) => {
    return {
      params: { slug: e }
    };
  })
  return { paths, fallback: false };
}

  const Event: React.FC<{slug: string}> = ({ slug }) => {
  const { data } = useSWR(`/api/events/${slug}`, (apiURL: string) => fetch(apiURL).then(res => res.json()));
  if (data === undefined) {
    return (
      <>
      <p>Loading</p>
      </>
    )
  }
  if (slug.startsWith('2022')) {
    const evtable = <EVTable ev={data.ev} matches={data.matches} team_sims={data.team_sims} />
    const matchtable = <MatchTable ev={data.ev} matches={data.matches} team_sims={data.team_sims} />
  } else {
    const evtable = <div></div>
    const matchtable = <div></div>
  }
  return (
    <>
      <p>This is the page for the event.</p>
      <Tab.Group>
        <Tab.List>
          <Tab className='p-4'>Rankings</Tab>
          <Tab className='p-4'>Matches</Tab>
        </Tab.List>
        <Tab.Panels>
          <Tab.Panel>
            <div className="grid grid-cols-2 break-after-column">
              <div className='col-span-2 lg:col-span-1'>
                <EVPlot data={data.ev} />
              </div>
              <div className='col-span-2 lg:col-span-1'>
                evtable
              </div>
            </div>
          </Tab.Panel>
          <Tab.Panel>
            matchtable
          </Tab.Panel>
        </Tab.Panels>
      </Tab.Group>
    </>
  )
}

const EventPage: React.FC<{slug: string, fallback: EventDataType}> = ({ slug, fallback }) => {
  return (
    <>
      <SWRConfig value={{ fallback }}>
        <Event slug={slug} />
      </SWRConfig>
    </>
  )
}

export default EventPage;