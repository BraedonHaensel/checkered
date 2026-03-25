import { Flag, Handshake, Home, PlusCircle, Trophy } from 'lucide-react'
import type { ReactNode } from 'react'

type Props = {
    onClick: () => void
}

type ButtonType = 'normal' | 'important' | 'danger' | 'accept'

// Search for opponent button
export const SearchButton = ({ onClick }: Props) => {
    return (
        <CustomButton onClick={onClick} type="important" hint="New Game">
            <PlusCircle />
        </CustomButton>
    )
}

export const LeaderboardButton = ({ onClick }: Props) => {
    return (
        <CustomButton onClick={onClick} type="normal" hint="Leaderboards">
            <Trophy />
        </CustomButton>
    )
}

export const HomeButton = ({ onClick }: Props) => {
    return (
        <CustomButton onClick={onClick} hint={'Home'}>
            <Home />
        </CustomButton>
    )
}

export const ForfeitButton = ({ onClick }: Props) => {
    return (
        <CustomButton onClick={onClick} type={'danger'} hint={'Forfeit'}>
            <Flag />
        </CustomButton>
    )
}

export const DrawButton = ({
    onClick,
    requested,
}: Props & { requested: boolean }) => {
    return (
        <CustomButton
            onClick={onClick}
            type={requested ? 'accept' : 'normal'}
            hint="Request Draw"
        >
            <Handshake />
        </CustomButton>
    )
}

const colors: Record<ButtonType, string> = {
    danger: 'bg-red-500',
    normal: 'bg-neutral-400',
    important: 'bg-purple-400',
    accept: 'bg-green-500',
}

const CustomButton = ({
    children,
    type = 'normal',
    onClick,
    hint,
}: {
    children: ReactNode
    type?: ButtonType
    onClick: () => void
    hint: string
}) => {
    return (
        <button
            onClick={onClick}
            data-hint={hint}
            className={`relative flex items-center justify-center hover:after:absolute hover:after:bottom-[calc(100%+10px)] hover:after:left-[50%] hover:after:block hover:after:w-max hover:after:transform-[translateX(-50%)] hover:after:rounded-md hover:after:bg-neutral-400 hover:after:p-1 hover:after:text-black hover:after:content-[attr(data-hint)] ${type && 'flex-grow'} ${colors[type]}`}
        >
            {children}
        </button>
    )
}
