import { json } from 'react-router-dom'

import dummyData01 from './dtos/01-dummy-network-data.json'

export async function loader() {
  return json([...dummyData01])
}
