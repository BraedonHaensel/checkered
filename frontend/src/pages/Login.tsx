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
        <div className='loginPageRoot'>
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
    )
}

export default Login
