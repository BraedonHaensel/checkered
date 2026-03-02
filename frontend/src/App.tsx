import { useState } from 'react'
import './App.css'
import Home from './pages/Home'
import { Page } from './enums'
import Game from './pages/Game'

const pages: Record<Page, React.ComponentType<{ setPage: (p: Page) => void }>> = {
    [Page.HOME]: Home,
    [Page.GAME]: Game,
}

function App() {
    const [page, setPage] = useState<Page>(Page.HOME)

    const PageComponent = pages[page]
  return (
    <>
            <PageComponent setPage={setPage} />
    </>
  )
}

export default App
