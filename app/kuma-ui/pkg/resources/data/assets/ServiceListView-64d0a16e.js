import{K as V}from"./index-52545d1d.js";import{d as z,a as l,o as t,b as r,w as a,e as o,p as x,f as n,t as m,c as _,F as d,q as k,V as R,E as T,v as g,_ as B}from"./index-d015481a.js";import{A as D}from"./AppCollection-47b71b41.js";import{E as I}from"./ErrorBlock-90874856.js";import{S as L}from"./StatusBadge-ed77f93c.js";import{S as N}from"./SummaryView-a7a171ba.js";import{T as A}from"./TextWithCopyButton-b83bb297.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-b301e74b.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-32f2be0a.js";import"./CopyButton-7634543f.js";const E=z({__name:"ServiceListView",setup(K){return($,q)=>{const h=l("RouteTitle"),u=l("RouterLink"),w=l("KCard"),C=l("RouterView"),S=l("AppView"),v=l("DataSource"),b=l("RouteView");return t(),r(v,{src:"/me"},{default:a(({data:y})=>[y?(t(),r(b,{key:0,name:"service-list-view",params:{page:1,size:y.pageSize,mesh:"",service:""}},{default:a(({route:s,t:p})=>[o(v,{src:`/meshes/${s.params.mesh}/service-insights?page=${s.params.page}&size=${s.params.size}`},{default:a(({data:i,error:c})=>[o(S,null,{title:a(()=>[x("h2",null,[o(h,{title:p("services.routes.items.title")},null,8,["title"])])]),default:a(()=>[n(),o(w,null,{default:a(()=>[c!==void 0?(t(),r(I,{key:0,error:c},null,8,["error"])):(t(),r(D,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":p("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":parseInt(s.params.page),"page-size":parseInt(s.params.size),total:i==null?void 0:i.total,items:i==null?void 0:i.items,error:c,"is-selected-row":e=>e.name===s.params.service,onChange:s.update},{name:a(({row:e})=>[o(u,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:s.params.page,size:s.params.size}}},{default:a(()=>[n(m(e.name),1)]),_:2},1032,["to"])]),serviceType:a(({rowValue:e})=>[n(m(e||"internal"),1)]),addressPort:a(({rowValue:e})=>[e?(t(),r(A,{key:0,text:e},null,8,["text"])):(t(),_(d,{key:1},[n(m(p("common.collection.none")),1)],64))]),online:a(({row:e})=>[e.dataplanes?(t(),_(d,{key:0},[n(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(t(),_(d,{key:1},[n(m(p("common.collection.none")),1)],64))]),status:a(({row:e})=>[o(L,{status:e.status||"not_available"},null,8,["status"])]),details:a(({row:e})=>[o(u,{class:"details-link","data-testid":"details-link",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:a(()=>[n(m(p("common.collection.details_link"))+" ",1),o(k(R),{display:"inline-block",decorative:"",size:k(V)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","is-selected-row","onChange"]))]),_:2},1024),n(),s.params.service?(t(),r(C,{key:0},{default:a(e=>[o(N,{onClose:f=>s.replace({name:"service-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[(t(),r(T(e.Component),{name:s.params.service,service:i==null?void 0:i.items.find(f=>f.name===s.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):g("",!0)]),_:1})}}});const M=B(E,[["__scopeId","data-v-81b3f1d1"]]);export{M as default};
