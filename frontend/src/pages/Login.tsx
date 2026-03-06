import { useRef } from 'react'
import { Page } from '../enums'

const Login = ({
    setUser,
    setPage,
}: {
    setUser: (user: string) => void
    setPage: (page: Page) => void
}) => {
    const usernameRef = useRef<HTMLInputElement>(null)
    return (
        <div className="loginPageRoot">
            <div className="loginContainer">
                <div className='loginTop'>
                    <h1 className="appTitle">CHECKERED</h1>
                    <p className="appSubtitle">
                      Online checkers game. Challenge friends and climb the ladder.
                    </p>
                </div>
                <div className='loginBottom'>
                    <div className='loginCard'>
                        <h2 className="loginTitle">USER LOGIN</h2>
                        <input type="text" placeholder="Username" ref={usernameRef} className="loginInput"/>
                        <button className="loginButton"
                            onClick={() => {
                                if (usernameRef.current?.value) {
                                    setUser(usernameRef.current.value || '')
                                    setPage(Page.HOME)
                                }
                            }}
                        >
                            Login
                        </button>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Login
