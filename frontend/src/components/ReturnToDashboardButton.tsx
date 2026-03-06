import { Page } from "../enums";

export const ReturnToDashboardButton = ({setPage}: {setPage: (page: Page) => void}) => {
    return <button onClick={() => setPage(Page.HOME)}>
        Return to Dashboard
    </button>
}
