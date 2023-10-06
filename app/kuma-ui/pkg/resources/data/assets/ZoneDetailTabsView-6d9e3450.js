import{d as N,$ as R,N as T,a0 as E,t as A,o as l,j as D,h as c,w as o,i as e,an as B,az as Z,l as _,C as h,am as M,g as d,m as y,k as I,r as b,E as $,s as L,a4 as j,a1 as P,n as G}from"./index-622cbb72.js";import{_ as Y}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-ef324a9f.js";import{N as q}from"./NavTabs-7056555a.js";const J=N({__name:"ZoneActionMenu",props:{zoneOverview:{type:Object,required:!0},kpopAttributes:{type:Object,default:()=>({placement:"bottomEnd"})}},setup(x){const s=x,{t:r}=R(),O=T(),z=E(),u=A(!1);function v(){u.value=!u.value}async function w(){await O.deleteZone({name:s.zoneOverview.name})}function t(){z.push({name:"zone-cp-list-view"})}return(i,n)=>(l(),D("div",null,[c(e(M),{"button-appearance":"creation","kpop-attributes":s.kpopAttributes,label:e(r)("zones.action_menu.toggle_button"),"show-caret":"",width:"280"},{items:o(()=>[c(e(B),{"is-dangerous":"","data-testid":"delete-button",onClick:Z(v,["prevent"])},{default:o(()=>[_(h(e(r)("zones.action_menu.delete_button")),1)]),_:1},8,["onClick"])]),_:1},8,["kpop-attributes","label"]),_(),u.value?(l(),d(Y,{key:0,"confirmation-text":s.zoneOverview.name,"delete-function":w,"is-visible":"","action-button-text":e(r)("common.delete_modal.proceed_button"),title:e(r)("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:v,onDelete:t},{"body-content":o(()=>[y("p",null,h(e(r)("common.delete_modal.text1",{type:"Zone",name:s.zoneOverview.name})),1),_(),y("p",null,h(e(r)("common.delete_modal.text2")),1)]),_:1},8,["confirmation-text","action-button-text","title"])):I("",!0)]))}}),H=N({__name:"ZoneDetailTabsView",setup(x){var w;const{t:s}=R(),z=(((w=E().getRoutes().find(t=>t.name==="zone-cp-detail-tabs-view"))==null?void 0:w.children)??[]).map(t=>{var a,p;const i=typeof t.name>"u"?(a=t.children)==null?void 0:a[0]:t,n=i.name,m=((p=i.meta)==null?void 0:p.module)??"";return{title:s(`zone-cps.routes.item.navigation.${n}`),routeName:n,module:m}}),u=A([]),v=t=>{var m,f;const i=[],n=((m=t.zoneInsight)==null?void 0:m.subscriptions)??[];if(n.length>0){const a=n[n.length-1],p=a.version.kumaCp.version||"-",{kumaCpGlobalCompatible:k=!0}=a.version.kumaCp;a.config&&((f=JSON.parse(a.config))==null?void 0:f.store.type)==="memory"&&i.push({kind:"ZONE_STORE_TYPE_MEMORY",payload:{}}),k||i.push({kind:"INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",payload:{zoneCpVersion:p}})}u.value=i};return(t,i)=>{const n=b("RouteTitle"),m=b("RouterView"),f=b("AppView"),a=b("DataSource"),p=b("RouteView");return l(),d(p,{name:"zone-cp-detail-tabs-view",params:{zone:""}},{default:o(({can:k,route:C})=>[c(a,{src:`/zone-cps/${C.params.zone}`,onChange:v},{default:o(({data:g,error:V})=>[V!==void 0?(l(),d($,{key:0,error:V},null,8,["error"])):g===void 0?(l(),d(L,{key:1})):(l(),d(f,{key:2,breadcrumbs:[{to:{name:"zone-cp-list-view"},text:e(s)("zone-cps.routes.item.breadcrumbs")}]},j({title:o(()=>[y("h1",null,[c(P,{text:C.params.zone},{default:o(()=>[c(n,{title:e(s)("zone-cps.routes.item.title",{name:C.params.zone}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:o(()=>[_(),_(),c(q,{class:"route-zone-detail-view-tabs",tabs:e(z)},null,8,["tabs"]),_(),c(m,null,{default:o(S=>[(l(),d(G(S.Component),{data:g,notifications:u.value},null,8,["data","notifications"]))]),_:2},1024)]),_:2},[k("create zones")?{name:"actions",fn:o(()=>[c(J,{"zone-overview":g},null,8,["zone-overview"])]),key:"0"}:void 0]),1032,["breadcrumbs"]))]),_:2},1032,["src"])]),_:1})}}});export{H as default};
