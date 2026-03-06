export const IngameDetails = ({children}: {children: any[] | any}) => {
    return <div className="p-5 w-full h-full">
        <div className="bg-neutral-900 w-full h-full rounded-lg grid grid-rows-[1fr_min-content]">
            <div>

            </div>
            <div className="flex flex-row flex-wrap justify-between p-2 gap-2">
                {children}
                {children}
            </div>
        </div>
    </div>
}
