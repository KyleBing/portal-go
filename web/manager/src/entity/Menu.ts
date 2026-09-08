enum EnumMenuType {
    '目录' = 1,
    '菜单',
    '按钮',
    '外链'
}
enum EnumMenuVisible {
    '可见' = 1,
    '已隐藏'
}


interface EntityMenu {
    name: string,
    type: EnumMenuType,
    path: string,
    match_path: string,
    component: string,
    visible: EnumMenuVisible,
    redirect: string,
    icon: string,
    isNeedAdminPermission: boolean,
    children: Array<EntityMenu>
}


export {
    type EntityMenu,
    EnumMenuType, EnumMenuVisible
}
