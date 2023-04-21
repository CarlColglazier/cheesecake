import Link from 'next/link'

export default function Home() {
  return (
    <>
      <p>Cheesecake is a scouting tool based on Bayesian statistics.</p>

      <p>See current <strong><Link href="/events" className='underline'>events</Link></strong>.</p>
    </>
  )
}
