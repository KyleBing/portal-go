import { request } from "./request";
import { Mob, Character, Plant, Thing, CookingRecipe, Log, Material, Craft, Coder, Version, CraftTab } from "../model/starve";


const BASE_URL = '/starve-new'
export default {
    // ==================== 列表接口 (GET) ====================
    
    // mob
    mobList(): Promise<Mob[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/mob/list`)},
    
    // character
    characterList(): Promise<Character[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/character/list`)},
    
    // plant
    plantList(): Promise<Plant[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/plant/list`)},
    
    // thing
    thingList(): Promise<Thing[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/thing/list`)},
    
    // cookingrecipe
    cookingrecipeList(): Promise<CookingRecipe[]> { return request('get', { all: 1 }, null, false, `${BASE_URL}/cookingrecipe/list`) },
    
    // material
    materialList(): Promise<Material[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/material/list`)},
    
    // craft
    craftList(): Promise<Craft[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/craft/list`)},
    
    // coder
    coderList(): Promise<Coder[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/coder/list`)},
    
    // log
    logList(): Promise<Log[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/log/list`)},

    // ==================== 详情接口 (GET) ====================
    
    // mob
    mobInfo(params: { id: number }): Promise<Mob> {return request('get', params, null, false, `${BASE_URL}/mob/info`)},
    
    // character
    characterInfo(params: { id: number }): Promise<Character> {return request('get', params, null, false, `${BASE_URL}/character/info`)},
    
    // plant
    plantInfo(params: { id: number }): Promise<Plant> {return request('get', params, null, false, `${BASE_URL}/plant/info`)},
    
    // thing
    thingInfo(params: { id: number }): Promise<Thing> {return request('get', params, null, false, `${BASE_URL}/thing/info`)},
    
    // cookingrecipe
    cookingrecipeInfo(params: { id: number }): Promise<CookingRecipe> {return request('get', params, null, false, `${BASE_URL}/cookingrecipe/info`)},
    
    // material
    materialInfo(params: { id: number }): Promise<Material> {return request('get', params, null, false, `${BASE_URL}/material/info`)},
    
    // craft
    craftInfo(params: { id: number }): Promise<Craft> {return request('get', params, null, false, `${BASE_URL}/craft/info`)},
    
    // coder
    coderInfo(params: { id: number }): Promise<Coder> {return request('get', params, null, false, `${BASE_URL}/coder/info`)},
    
    // log
    logInfo(params: { id: number }): Promise<Log> {return request('get', params, null, false, `${BASE_URL}/log/info`)},

    // ==================== 增删改接口 ====================
    
    // log
    logAdd(requestData: Log): Promise<Log> {return request('post', {}, requestData, false, `${BASE_URL}/log/add`)},
    logModify(requestData: Log): Promise<Log> {return request('put', {}, requestData, false, `${BASE_URL}/log/modify`)},
    logDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/log/delete`)},
    
    // character
    characterAdd(requestData: Character): Promise<Character> {return request('post', {}, requestData, false, `${BASE_URL}/character/add`)},
    characterModify(requestData: Character): Promise<Character> {return request('put', {}, requestData, false, `${BASE_URL}/character/modify`)},
    characterDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/character/delete`)},
    
    // mob
    mobAdd(requestData: Mob): Promise<Mob> {return request('post', {}, requestData, false, `${BASE_URL}/mob/add`)},
    mobModify(requestData: Mob): Promise<Mob> {return request('put', {}, requestData, false, `${BASE_URL}/mob/modify`)},
    mobDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/mob/delete`)},
    
    // plant
    plantAdd(requestData: Plant): Promise<Plant> {return request('post', {}, requestData, false, `${BASE_URL}/plant/add`)},
    plantModify(requestData: Plant): Promise<Plant> {return request('put', {}, requestData, false, `${BASE_URL}/plant/modify`)},
    plantDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/plant/delete`)},
    
    // thing
    thingAdd(requestData: Thing): Promise<Thing> {return request('post', {}, requestData, false, `${BASE_URL}/thing/add`)},
    thingModify(requestData: Thing): Promise<Thing> {return request('put', {}, requestData, false, `${BASE_URL}/thing/modify`)},
    thingDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/thing/delete`)},
    
    // material
    materialAdd(requestData: Material): Promise<Material> {return request('post', {}, requestData, false, `${BASE_URL}/material/add`)},
    materialModify(requestData: Material): Promise<Material> {return request('put', {}, requestData, false, `${BASE_URL}/material/modify`)},
    materialDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/material/delete`)},
    
    // craft
    craftAdd(requestData: Craft): Promise<Craft> {return request('post', {}, requestData, false, `${BASE_URL}/craft/add`)},
    craftModify(requestData: Craft): Promise<Craft> {return request('put', {}, requestData, false, `${BASE_URL}/craft/modify`)},
    craftDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/craft/delete`)},
    
    // cookingrecipe
    cookingrecipeAdd(requestData: CookingRecipe): Promise<CookingRecipe> {return request('post', {}, requestData, false, `${BASE_URL}/cookingrecipe/add`)},
    cookingrecipeModify(requestData: CookingRecipe): Promise<CookingRecipe> {return request('put', {}, requestData, false, `${BASE_URL}/cookingrecipe/modify`)},
    cookingrecipeDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/cookingrecipe/delete`)},
    
    // coder
    coderAdd(requestData: Coder): Promise<Coder> {return request('post', {}, requestData, false, `${BASE_URL}/coder/add`)},
    coderModify(requestData: Coder): Promise<Coder> {return request('put', {}, requestData, false, `${BASE_URL}/coder/modify`)},
    coderDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/coder/delete`)},

    // ==================== Version 管理接口 ====================
    
    // version
    versionList(): Promise<Version[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/version/list`)},
    versionInfo(params: { id: number }): Promise<Version> {return request('get', params, null, false, `${BASE_URL}/version/info`)},
    versionAdd(requestData: Version): Promise<Version> {return request('post', {}, requestData, false, `${BASE_URL}/version/add`)},
    versionModify(requestData: Version): Promise<Version> {return request('put', {}, requestData, false, `${BASE_URL}/version/modify`)},
    versionDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/version/delete`)},

    // ==================== CraftTab 管理接口 ====================
    
    // craftTab
    craftTabList(): Promise<CraftTab[]> {return request('get', { all: 1 }, null, false, `${BASE_URL}/craft-tab/list`)},
    craftTabInfo(params: { id: number }): Promise<CraftTab> {return request('get', params, null, false, `${BASE_URL}/craft-tab/info`)},
    craftTabAdd(requestData: CraftTab): Promise<CraftTab> {return request('post', {}, requestData, false, `${BASE_URL}/craft-tab/add`)},
    craftTabModify(requestData: CraftTab): Promise<CraftTab> {return request('put', {}, requestData, false, `${BASE_URL}/craft-tab/modify`)},
    craftTabDelete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, `${BASE_URL}/craft-tab/delete`)},
}