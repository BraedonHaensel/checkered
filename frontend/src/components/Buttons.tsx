import { Flag, Handshake, Home, PlusCircle, Trophy } from "lucide-react"

type Props = {
    onClick: () => void
}

type ButtonType = 'normal' | 'important' | 'danger'

// Search for opponent button
export const SearchButton = ({ onClick }: Props) => {
    return <CustomButton onClick={onClick} type="important" hint="New Game">
        <PlusCircle />
    </CustomButton>
}

export const LeaderboardButton = ({ onClick }: Props) => {
    return <CustomButton onClick={onClick} type="normal" hint="Leaderboards">
        <Trophy />
    </CustomButton>
}

export const DashboardButton = ({ onClick, exitsGame }: Props & {exitsGame: boolean}) => {
    return <CustomButton onClick={onClick} type={exitsGame ? "danger" : "normal"} hint={exitsGame ? "Forfeit" : "Home"}>
        {exitsGame ? <Flag /> : <Home />}
    </CustomButton>
}

export const DrawButton = ({onClick}: Props) => {
    return <CustomButton onClick={onClick} type="normal" hint="Request Draw">
        <Handshake />
    </CustomButton>
}

const colors: Record<ButtonType, string> = {
    danger: "red-500",
    normal: "neutral-400",
    important: "purple-400",
}

const CustomButton = ({children, type = 'normal', onClick, hint}: {children: any, type: 'normal' | 'important' | 'danger', onClick: () => void, hint: string}) => {
    return <button onClick={onClick} data-hint={hint} className={`relative hover:after:content-[attr(data-hint)] hover:after:absolute hover:after:bottom-[calc(100%+10px)] hover:after:left-[50%] hover:after:transform-[translateX(-50%)] hover:after:bg-neutral-400 hover:after:text-black hover:after:rounded-md hover:after:p-1 hover:after:w-max hover:after:block flex justify-center items-center ${type && "flex-grow"} bg-${colors[type]}`}>{children}</button>
    
}
