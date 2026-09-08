import {EntityMenu, EnumMenuType, EnumMenuVisible} from "@/entity/Menu.ts";

const MENUS_PRESET: Array<EntityMenu> = [
    {
        "name": "日记",
        "type": EnumMenuType['目录'],
        "path": "/diary",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "/diary/category",
        "icon": "Notebook",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "统计",
                "type": EnumMenuType['菜单'],
                "path": "/diary/statistic",
                "match_path": "",
                "component": "Statistics/StatisticMain.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "PieChart",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "日记类别管理",
                "type": EnumMenuType['菜单'],
                "path": "/diary/category",
                "match_path": "",
                "component": "Diary/DiaryCategory.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Folder",
                "isNeedAdminPermission": true,
                "children": []
            }

        ]
    },
    {
        "name": "五笔",
        "type": EnumMenuType['目录'],
        "path": "/wubi",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "/wubi/category",
        "icon": "Document",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "五笔类别",
                "type": EnumMenuType['菜单'],
                "path": "/wubi/category",
                "match_path": "",
                "component": "Wubi/WubiCategory.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Folder",
                "isNeedAdminPermission": true,
                "children": []
            },
            {
                "name": "五笔词条",
                "type": EnumMenuType['菜单'],
                "path": "/wubi/words",
                "match_path": "",
                "component": "Wubi/WubiWords.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Document",
                "isNeedAdminPermission": false,
                "children": []
            },
        ]
    },
    {
        "name": "二维码",
        "type": EnumMenuType['目录'],
        "path": "/qr",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "",
        "icon": "Picture",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "二维码管理",
                "type": EnumMenuType['菜单'],
                "path": "/qr/list",
                "match_path": "",
                "component": "Qr/QR.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Picture",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "点赞管理",
                "type": EnumMenuType['菜单'],
                "path": "/qr/thumbsup",
                "match_path": "",
                "component": "ThumbsUp/ThumbsUp.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Star",
                "isNeedAdminPermission": true,
                "children": []
            }
        ]
    },
    {
        "name": "图片与文件",
        "type": EnumMenuType['目录'],
        "path": "/file",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "",
        "icon": "Setting",
        "isNeedAdminPermission": true,
        "children": [
            {
                "name": "七牛云图床",
                "type": EnumMenuType['菜单'],
                "path": "/file/image-management",
                "match_path": "",
                "component": "Image/ImageManagement.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Picture",
                "isNeedAdminPermission": true,
                "children": []
            },
            {
                "name": "文件分享",
                "type": EnumMenuType['菜单'],
                "path": "/file/file-management",
                "match_path": "",
                "component": "File/FileManager.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Folder",
                "isNeedAdminPermission": true,
                "children": []
            }
        ]
    },
    {
        "name": "饥荒",
        "type": EnumMenuType['目录'],
        "path": "/starve",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "/starve/mob",
        "icon": "Food",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "日志管理",
                "type": EnumMenuType['菜单'],
                "path": "/starve/log",
                "match_path": "",
                "component": "Starve/LogList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Document",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "角色列表",
                "type": EnumMenuType['菜单'],
                "path": "/starve/character",
                "match_path": "",
                "component": "Starve/CharacterList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Avatar",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "生物列表",
                "type": EnumMenuType['菜单'],
                "path": "/starve/mob",
                "match_path": "",
                "component": "Starve/MobList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "User",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "植物列表",
                "type": EnumMenuType['菜单'],
                "path": "/starve/plant",
                "match_path": "",
                "component": "Starve/PlantList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Cherry",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "物品列表",
                "type": EnumMenuType['菜单'],
                "path": "/starve/thing",
                "match_path": "",
                "component": "Starve/ThingList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Box",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "烹饪食谱",
                "type": EnumMenuType['菜单'],
                "path": "/starve/cooking-recipe",
                "match_path": "",
                "component": "Starve/CookingRecipeList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Food",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "制作列表",
                "type": EnumMenuType['菜单'],
                "path": "/starve/craft",
                "match_path": "",
                "component": "Starve/CraftList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Tools",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "代码管理",
                "type": EnumMenuType['菜单'],
                "path": "/starve/coder",
                "match_path": "",
                "component": "Starve/CoderList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Document",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "版本管理",
                "type": EnumMenuType['菜单'],
                "path": "/starve/version",
                "match_path": "",
                "component": "Starve/VersionList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Collection",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "制作标签管理",
                "type": EnumMenuType['菜单'],
                "path": "/starve/craft-tab",
                "match_path": "",
                "component": "Starve/CraftTabList.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "PriceTag",
                "isNeedAdminPermission": false,
                "children": []
            }
        ]
    },
    {
        "name": "Apple",
        "type": EnumMenuType['目录'],
        "path": "/apple",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "/apple/chip-a",
        "icon": "Apple",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "A 系列芯片",
                "type": EnumMenuType['菜单'],
                "path": "/apple/chip-a",
                "match_path": "",
                "component": "Apple/AppleChipA.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Cpu",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "S 系列芯片",
                "type": EnumMenuType['菜单'],
                "path": "/apple/chip-s",
                "match_path": "",
                "component": "Apple/AppleChipS.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Watch",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "M 系列芯片",
                "type": EnumMenuType['菜单'],
                "path": "/apple/chip-m",
                "match_path": "",
                "component": "Apple/AppleChipM.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Monitor",
                "isNeedAdminPermission": false,
                "children": []
            }
        ]
    },
    {
        "name": "系统配置",
        "type": EnumMenuType['目录'],
        "path": "/system",
        "match_path": "",
        "component": "",
        "visible": EnumMenuVisible['可见'],
        "redirect": "",
        "icon": "Setting",
        "isNeedAdminPermission": false,
        "children": [
            {
                "name": "用户管理",
                "type": EnumMenuType['菜单'],
                "path": "/system/permission-setup/user-list",
                "match_path": "",
                "component": "User/User.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "User",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "邮件",
                "type": EnumMenuType['菜单'],
                "path": "/system/email",
                "match_path": "",
                "component": "Mail.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Message",
                "isNeedAdminPermission": true,
                "children": []
            },
            {
                "name": "修改密码",
                "type": EnumMenuType['菜单'],
                "path": "/system/change-password",
                "match_path": "",
                "component": "ChangePassword.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "Key",
                "isNeedAdminPermission": false,
                "children": []
            },
            {
                "name": "关于",
                "type": EnumMenuType['菜单'],
                "path": "/system/about",
                "match_path": "",
                "component": "About.vue",
                "visible": EnumMenuVisible['可见'],
                "redirect": "",
                "icon": "InfoFilled",
                "isNeedAdminPermission": false,
                "children": []
            },
        ]
    }
]

export {
    MENUS_PRESET
}
