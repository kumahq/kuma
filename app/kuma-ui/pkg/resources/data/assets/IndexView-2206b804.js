import{d as N,g as R,R as T,a4 as E,y as A,o as l,l as B,j as c,w as o,k as e,a3 as D,aM as M,n as _,H as g,a1 as I,i as d,p as h,m as Z,r as b,E as $,x as L,a8 as j,a5 as P,q}from"./index-21079cd9.js";import{_ as G}from"./DeleteResourceModal.vue_vue_type_script_setup_true_lang-f47d88f9.js";import{N as Y}from"./NavTabs-ddecd1ce.js";const H=N({__name:"ZoneActionMenu",props:{zoneOverview:{type:Object,required:!0},kpopAttributes:{type:Object,default:()=>({placement:"bottomEnd"})}},setup(x){const s=x,{t:r}=R(),O=T(),z=E(),u=A(!1);function v(){u.value=!u.value}async function w(){await O.deleteZone({name:s.zoneOverview.name})}function t(){z.push({name:"zone-cp-list-view"})}return(i,n)=>(l(),B("div",null,[c(e(I),{"button-appearance":"creation","kpop-attributes":s.kpopAttributes,label:e(r)("zones.action_menu.toggle_button"),"show-caret":"",width:"280"},{items:o(()=>[c(e(D),{"is-dangerous":"","data-testid":"delete-button",onClick:M(v,["prevent"])},{default:o(()=>[_(g(e(r)("zones.action_menu.delete_button")),1)]),_:1},8,["onClick"])]),_:1},8,["kpop-attributes","label"]),_(),u.value?(l(),d(G,{key:0,"confirmation-text":s.zoneOverview.name,"delete-function":w,"is-visible":"","action-button-text":e(r)("common.delete_modal.proceed_button"),title:e(r)("common.delete_modal.title",{type:"Zone"}),"data-testid":"delete-zone-modal",onCancel:v,onDelete:t},{"body-content":o(()=>[h("p",null,g(e(r)("common.delete_modal.text1",{type:"Zone",name:s.zoneOverview.name})),1),_(),h("p",null,g(e(r)("common.delete_modal.text2")),1)]),_:1},8,["confirmation-text","action-button-text","title"])):Z("",!0)]))}}),F=N({__name:"IndexView",setup(x){var w;const{t:s}=R(),z=(((w=E().getRoutes().find(t=>t.name==="zone-cp-detail-tabs-view"))==null?void 0:w.children)??[]).map(t=>{var a,p;const i=typeof t.name>"u"?(a=t.children)==null?void 0:a[0]:t,n=i.name,m=((p=i.meta)==null?void 0:p.module)??"";return{title:s(`zone-cps.routes.item.navigation.${n}`),routeName:n,module:m}}),u=A([]),v=t=>{var m,f;const i=[],n=((m=t.zoneInsight)==null?void 0:m.subscriptions)??[];if(n.length>0){const a=n[n.length-1],p=a.version.kumaCp.version||"-",{kumaCpGlobalCompatible:k=!0}=a.version.kumaCp;a.config&&((f=JSON.parse(a.config))==null?void 0:f.store.type)==="memory"&&i.push({kind:"ZONE_STORE_TYPE_MEMORY",payload:{}}),k||i.push({kind:"INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS",payload:{zoneCpVersion:p}})}u.value=i};return(t,i)=>{const n=b("RouteTitle"),m=b("RouterView"),f=b("AppView"),a=b("DataSource"),p=b("RouteView");return l(),d(p,{name:"zone-cp-detail-tabs-view",params:{zone:""}},{default:o(({can:k,route:y})=>[c(a,{src:`/zone-cps/${y.params.zone}`,onChange:v},{default:o(({data:C,error:V})=>[V!==void 0?(l(),d($,{key:0,error:V},null,8,["error"])):C===void 0?(l(),d(L,{key:1})):(l(),d(f,{key:2,breadcrumbs:[{to:{name:"zone-cp-list-view"},text:e(s)("zone-cps.routes.item.breadcrumbs")}]},j({title:o(()=>[h("h1",null,[c(P,{text:y.params.zone},{default:o(()=>[c(n,{title:e(s)("zone-cps.routes.item.title",{name:y.params.zone}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:o(()=>[_(),_(),c(Y,{class:"route-zone-detail-view-tabs",tabs:e(z)},null,8,["tabs"]),_(),c(m,null,{default:o(S=>[(l(),d(q(S.Component),{data:C,notifications:u.value},null,8,["data","notifications"]))]),_:2},1024)]),_:2},[k("create zones")?{name:"actions",fn:o(()=>[c(H,{"zone-overview":C},null,8,["zone-overview"])]),key:"0"}:void 0]),1032,["breadcrumbs"]))]),_:2},1032,["src"])]),_:1})}}});export{F as default};
