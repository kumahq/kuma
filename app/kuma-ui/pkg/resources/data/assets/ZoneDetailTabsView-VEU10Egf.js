import{d as O,l as N,N as B,V as R,B as A,a as i,o as r,c as I,e as s,w as o,aD as S,q as n,t as k,f as u,b as m,m as h,p as Z,C as M,a1 as K,aA as D}from"./index-WjsS4EhC.js";import{_ as $}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-CqEXSzEA.js";import{E as L}from"./ErrorBlock-9G-DlLfx.js";import{_ as P}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-GTOSCqjY.js";import{N as j}from"./NavTabs-11hgTURx.js";import{T as q}from"./TextWithCopyButton-bzYK7m0g.js";import"./index-FZCiQto1.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-jUAJBWzU.js";import"./CopyButton-11uSqZIk.js";const G=O({__name:"ZoneActionMenu",props:{zoneOverview:{type:Object,required:!0},kpopAttributes:{type:Object,default:()=>({placement:"bottomEnd"})}},setup(C){const{t}=N(),x=B(),y=R(),c=C,l=A(!1);function _(){l.value=!l.value}async function f(){await x.deleteZone({name:c.zoneOverview.name})}function e(){y.push({name:"zone-cp-list-view"})}return(a,p)=>{const d=i("KDropdownItem"),v=i("KDropdown");return r(),I("div",null,[s(v,{"kpop-attributes":c.kpopAttributes,"trigger-text":n(t)("zones.action_menu.toggle_button"),"show-caret":"",width:"280"},{items:o(()=>[s(d,{danger:"","data-testid":"delete-button",onClick:S(_,["prevent"])},{default:o(()=>[u(k(n(t)("zones.action_menu.delete_button")),1)]),_:1})]),_:1},8,["kpop-attributes","trigger-text"]),u(),l.value?(r(),m($,{key:0,"confirmation-text":c.zoneOverview.name,"delete-function":f,"is-visible":"","action-button-text":n(t)("common.delete_modal.proceed_button"),title:n(t)("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:_,onDelete:e},{default:o(()=>[h("p",null,k(n(t)("common.delete_modal.text1",{type:"Zone",name:c.zoneOverview.name})),1),u(),h("p",null,k(n(t)("common.delete_modal.text2")),1)]),_:1},8,["confirmation-text","action-button-text","title"])):Z("",!0)])}}}),te=O({__name:"ZoneDetailTabsView",setup(C){var f;const{t}=N(),c=(((f=R().getRoutes().find(e=>e.name==="zone-cp-detail-tabs-view"))==null?void 0:f.children)??[]).map(e=>{var b,w;const a=typeof e.name>"u"?(b=e.children)==null?void 0:b[0]:e,p=a.name,d=((w=a.meta)==null?void 0:w.module)??"";return{title:t(`zone-cps.routes.item.navigation.${p}`),routeName:p,module:d}}),l=A([]),_=e=>{const a=[];e.zoneInsight.store==="memory"&&a.push({kind:"ZONE_STORE_TYPE_MEMORY",payload:{}}),D(e.zoneInsight,"version.kumaCp.kumaCpGlobalCompatible","true")||a.push({kind:"INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",payload:{zoneCpVersion:D(e.zoneInsight,"version.kumaCp.version",t("common.collection.none"))}}),l.value=a};return(e,a)=>{const p=i("RouteTitle"),d=i("RouterView"),v=i("AppView"),b=i("DataSource"),w=i("RouteView");return r(),m(w,{name:"zone-cp-detail-tabs-view",params:{zone:""}},{default:o(({can:E,route:z})=>[s(b,{src:`/zone-cps/${z.params.zone}`,onChange:_},{default:o(({data:g,error:V})=>[V!==void 0?(r(),m(L,{key:0,error:V},null,8,["error"])):g===void 0?(r(),m(P,{key:1})):(r(),m(v,{key:2,breadcrumbs:[{to:{name:"zone-cp-list-view"},text:n(t)("zone-cps.routes.item.breadcrumbs")}]},K({title:o(()=>[h("h1",null,[s(q,{text:z.params.zone},{default:o(()=>[s(p,{title:n(t)("zone-cps.routes.item.title",{name:z.params.zone})},null,8,["title"])]),_:2},1032,["text"])])]),default:o(()=>[u(),u(),s(j,{class:"route-zone-detail-view-tabs",tabs:n(c)},null,8,["tabs"]),u(),s(d,null,{default:o(T=>[(r(),m(M(T.Component),{data:g,notifications:l.value},null,8,["data","notifications"]))]),_:2},1024)]),_:2},[E("create zones")?{name:"actions",fn:o(()=>[s(G,{"zone-overview":g},null,8,["zone-overview"])]),key:"0"}:void 0]),1032,["breadcrumbs"]))]),_:2},1032,["src"])]),_:1})}}});export{te as default};
