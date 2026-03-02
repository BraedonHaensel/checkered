import { useEffect, useState } from 'react'
import './App.css'
import Home from './pages/Home'
import { Page } from './enums'
import Game from './pages/Game'
import { Backend } from './api/backend'

const pages: Record<Page, React.ComponentType<{ setPage: (p: Page) => void }>> = {
    [Page.HOME]: Home,
    [Page.GAME]: Game,
}

function App() {
    const [page, setPage] = useState<Page>(Page.HOME)

    useEffect(() => {
        Backend.instance().get("get_leaderboard", {}).then(res => {
            console.log(res)
        }).catch(err => {
            console.error("Failed to fetch leaderboard: ", err);
        });
    })

    const PageComponent = pages[page]
    return (
        <>
            <PageComponent setPage={setPage} />
        </>
    )
}

export default App
