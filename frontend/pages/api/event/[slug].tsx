import path from 'path';
import { promises as fs } from 'fs';

export default async function handler(req: any, res: { status: (arg0: number) => { (): any; new(): any; json: { (arg0: string): void; new(): any; }; }; }) {
  const { slug } = req.query
  const jsonDirectory = path.join(process.cwd(), '../files/api/event/');
  const fileContents = await fs.readFile(jsonDirectory + `${slug}.json`, 'utf8')
  res.status(200).json(JSON.parse(fileContents));
}