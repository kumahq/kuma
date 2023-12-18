import{d as V,l as N,u as I,W as R,m as E,a as i,o as r,c as B,e as s,w as o,aC as S,f as u,t as k,q as n,b as m,p as h,s as Z,Y as M,D as K,az as O}from"./index-7a0947c2.js";import{_ as $}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-15e78222.js";import{E as L}from"./ErrorBlock-78880c60.js";import{_ as P}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-a6d76488.js";import{N as Y}from"./NavTabs-e9c664ed.js";import{T as j}from"./TextWithCopyButton-3aa03737.js";import"./index-fce48c05.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-1c689249.js";import"./CopyButton-a5c25cdd.js";const q=V({__name:"ZoneActionMenu",props:{zoneOverview:{type:Object,required:!0},kpopAttributes:{type:Object,default:()=>({placement:"bottomEnd"})}},setup(y){const{t}=N(),C=I(),x=R(),c=y,l=E(!1);function _(){l.value=!l.value}async function f(){await C.deleteZone({name:c.zoneOverview.name})}function e(){x.push({name:"zone-cp-list-view"})}return(a,p)=>{const d=i("KDropdownItem"),v=i("KDropdown");return r(),B("div",null,[s(v,{"kpop-attributes":c.kpopAttributes,"trigger-text":n(t)("zones.action_menu.toggle_button"),"show-caret":"",width:"280"},{items:o(()=>[s(d,{danger:"","data-testid":"delete-button",onClick:S(_,["prevent"])},{default:o(()=>[u(k(n(t)("zones.action_menu.delete_button")),1)]),_:1})]),_:1},8,["kpop-attributes","trigger-text"]),u(),l.value?(r(),m($,{key:0,"confirmation-text":c.zoneOverview.name,"delete-function":f,"is-visible":"","action-button-text":n(t)("common.delete_modal.proceed_button"),title:n(t)("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:_,onDelete:e},{"body-content":o(()=>[h("p",null,k(n(t)("common.delete_modal.text1",{type:"Zone",name:c.zoneOverview.name})),1),u(),h("p",null,k(n(t)("common.delete_modal.text2")),1)]),_:1},8,["confirmation-text","action-button-text","title"])):Z("",!0)])}}}),te=V({__name:"ZoneDetailTabsView",setup(y){var f;const{t}=N(),c=(((f=R().getRoutes().find(e=>e.name==="zone-cp-detail-tabs-view"))==null?void 0:f.children)??[]).map(e=>{var b,w;const a=typeof e.name>"u"?(b=e.children)==null?void 0:b[0]:e,p=a.name,d=((w=a.meta)==null?void 0:w.module)??"";return{title:t(`zone-cps.routes.item.navigation.${p}`),routeName:p,module:d}}),l=E([]),_=e=>{const a=[];e.zoneInsight.store==="memory"&&a.push({kind:"ZONE_STORE_TYPE_MEMORY",payload:{}}),O(e.zoneInsight,"version.kumaCp.kumaCpGlobalCompatible","true")||a.push({kind:"INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",payload:{zoneCpVersion:O(e.zoneInsight,"version.kumaCp.version",t("common.collection.none"))}}),l.value=a};return(e,a)=>{const p=i("RouteTitle"),d=i("RouterView"),v=i("AppView"),b=i("DataSource"),w=i("RouteView");return r(),m(w,{name:"zone-cp-detail-tabs-view",params:{zone:""}},{default:o(({can:T,route:z})=>[s(b,{src:`/zone-cps/${z.params.zone}`,onChange:_},{default:o(({data:g,error:D})=>[D!==void 0?(r(),m(L,{key:0,error:D},null,8,["error"])):g===void 0?(r(),m(P,{key:1})):(r(),m(v,{key:2,breadcrumbs:[{to:{name:"zone-cp-list-view"},text:n(t)("zone-cps.routes.item.breadcrumbs")}]},M({title:o(()=>[h("h1",null,[s(j,{text:z.params.zone},{default:o(()=>[s(p,{title:n(t)("zone-cps.routes.item.title",{name:z.params.zone})},null,8,["title"])]),_:2},1032,["text"])])]),default:o(()=>[u(),u(),s(Y,{class:"route-zone-detail-view-tabs",tabs:n(c)},null,8,["tabs"]),u(),s(d,null,{default:o(A=>[(r(),m(K(A.Component),{data:g,notifications:l.value},null,8,["data","notifications"]))]),_:2},1024)]),_:2},[T("create zones")?{name:"actions",fn:o(()=>[s(q,{"zone-overview":g},null,8,["zone-overview"])]),key:"0"}:void 0]),1032,["breadcrumbs"]))]),_:2},1032,["src"])]),_:1})}}});export{te as default};
