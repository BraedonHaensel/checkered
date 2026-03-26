import { useState, useEffect } from 'react'
import Home from './pages/Home'
import { Page } from './enums'
import Game from './pages/Game'
import Login from './pages/Login'
import Leaderboard from './pages/Leaderboard'

const pages: Record<
    Page,
    React.ComponentType<{
        setPage: (p: Page) => void
        username: string
        setUsername: (user: string) => void
    }>
> = {
    [Page.HOME]: Home,
    [Page.GAME]: Game,
    [Page.LOGIN]: Login,
    [Page.LEADERBOARD]: Leaderboard,
}

function App() {
    const [username, setUsername] = useState<string>('')
    const [page, setPage] = useState<Page>(Page.LOGIN)

    // Adding body style speficially for the login page - just for the background
    useEffect(() => {
        if (page === Page.LOGIN) {
            document.body.classList.add('login-page')
        } else {
            document.body.classList.remove('login-page')
        }

        // Cleanup
        return () => {
            document.body.classList.remove('login-page')
        }
    }, [page])

    const PageComponent = pages[page]
    return (
        <>
            <PageComponent setPage={setPage} username={username} setUsername={setUsername} />
        </>
    )
}

export default App
