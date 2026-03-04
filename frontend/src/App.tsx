import { useState } from 'react'
import './App.css'
import Home from './pages/Home'
import { Page } from './enums'
import Game from './pages/Game'
import Login from './pages/Login'
import Leaderboard from './pages/Leaderboard'

const pages: Record<
    Page,
    React.ComponentType<{
        setPage: (p: Page) => void
        user: string
        setUser: (user: string) => void
    }>
> = {
    [Page.HOME]: Home,
    [Page.GAME]: Game,
    [Page.LOGIN]: Login,
    [Page.LEADERBOARD]: Leaderboard,
}

function App() {
    const [user, setUser] = useState<string>('')
    const [page, setPage] = useState<Page>(Page.LOGIN)

    const PageComponent = pages[page]
    return (
        <>
            <PageComponent setPage={setPage} user={user} setUser={setUser} />
        </>
    )
}

export default App
