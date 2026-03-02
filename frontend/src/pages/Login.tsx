import { useRef } from "react"
import { Page } from "../enums"

const Login = ({setUser, setPage}: {setUser: (user: string) => void, setPage: (page: Page) => void}) => {
    const usernameRef = useRef<HTMLInputElement>(null)
    return <div>
        <input type="text" placeholder="Username" ref={usernameRef} />
        <button onClick={() => {
            if(usernameRef.current?.value) {
                setUser(usernameRef.current.value || "")
                setPage(Page.HOME);
            }
        }}>Login</button>
    </div>
}

export default Login;
