import { PlusCircle } from "lucide-react"

type Props = {
    onClick: () => void
}

// Search for opponent button
const SearchButton = ({ onClick }: Props) => {
    return <CustomButton onClick={onClick} highlight={true}>
        <PlusCircle />
    </CustomButton>
}

const CustomButton = ({children, highlight = false, onClick}: {children: any, highlight: boolean, onClick: () => void}) => {
    const _ = highlight;
    return <button onClick={onClick} className={`relative hover:after:content-['Find_New_Game'] hover:after:absolute hover:after:bottom-[calc(100%+10px)] hover:after:left-[50%] hover:after:transform-[translateX(-50%)] hover:after:bg-neutral-400 hover:after:text-black hover:after:rounded-md hover:after:p-1 hover:after:w-max hover:after:block ${highlight && "flex-grow"} ${highlight && "bg-purple-400"}`}>{children}</button>
    
}

export default SearchButton
