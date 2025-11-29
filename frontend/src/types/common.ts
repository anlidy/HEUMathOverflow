export interface Response<T> {
    code: number
    message: string
    data: T
}

export interface NavItem {
    label: string
    path: string
}

export interface TopicItem {
    id: number
    name: string
}
