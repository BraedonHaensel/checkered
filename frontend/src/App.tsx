import { useEffect, useState } from 'react'
import './App.css'
import Home from './pages/Home'
import { Page } from './enums'
import Game from './pages/Game'
import { Backend } from './api/backend'
import Login from './pages/Login'

const pages: Record<Page, React.ComponentType<{ setPage: (p: Page) => void, user: string, setUser: (user: string) => void }>> = {
    [Page.HOME]: Home,
    [Page.GAME]: Game,
    [Page.LOGIN]: Login,
}

function App() {
    const [user, setUser] = useState<string>("")
    const [page, setPage] = useState<Page>(Page.LOGIN)

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
            <PageComponent setPage={setPage} user={user} setUser={setUser} />
        </>
    )
}

export default App
