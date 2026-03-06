export const IngameDetails = ({children, statusMessage}: {children: any[] | any, statusMessage: string}) => {
    return <div className="p-5 w-full h-full">
        <div className="bg-neutral-900 w-full h-full rounded-lg grid grid-rows-[1fr_min-content]">
            <div className="p-2 flex flex-col">
                <h2 className="w-full text-center text-xl">{statusMessage}</h2>
            </div>
            <div className="flex flex-row flex-wrap justify-between p-2 gap-2">
                {children}
            </div>
        </div>
    </div>
}
