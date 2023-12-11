import{K as b}from"./index-52545d1d.js";import{d as x,a as l,o as t,b as r,w as s,e as o,p as R,f as n,t as m,c as _,F as d,q as k,U as T,D as B,s as g,_ as D}from"./index-8567ed34.js";import{A as I}from"./AppCollection-668a360f.js";import{E as L}from"./ErrorBlock-3040d559.js";import{S as N}from"./StatusBadge-bb50b00b.js";import{S as A}from"./SummaryView-364e16b7.js";import{T as h}from"./TextWithCopyButton-cadc290c.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-c8b34455.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-184215be.js";import"./CopyButton-faeef64d.js";const E=x({__name:"ServiceListView",setup(K){return($,q)=>{const w=l("RouteTitle"),u=l("RouterLink"),C=l("KCard"),S=l("RouterView"),z=l("AppView"),v=l("DataSource"),V=l("RouteView");return t(),r(v,{src:"/me"},{default:s(({data:y})=>[y?(t(),r(V,{key:0,name:"service-list-view",params:{page:1,size:y.pageSize,mesh:"",service:""}},{default:s(({route:a,t:c})=>[o(v,{src:`/meshes/${a.params.mesh}/service-insights?page=${a.params.page}&size=${a.params.size}`},{default:s(({data:i,error:p})=>[o(z,null,{title:s(()=>[R("h2",null,[o(w,{title:c("services.routes.items.title")},null,8,["title"])])]),default:s(()=>[n(),o(C,null,{default:s(()=>[p!==void 0?(t(),r(L,{key:0,error:p},null,8,["error"])):(t(),r(I,{key:1,class:"service-collection","data-testid":"service-collection","empty-state-message":c("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"DP proxies (online / total)",key:"online"},{label:"Status",key:"status"},{label:"Details",key:"details",hideLabel:!0}],"page-number":parseInt(a.params.page),"page-size":parseInt(a.params.size),total:i==null?void 0:i.total,items:i==null?void 0:i.items,error:p,"is-selected-row":e=>e.name===a.params.service,onChange:a.update},{name:s(({row:e})=>[o(h,{text:e.name},{default:s(()=>[o(u,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name},query:{page:a.params.page,size:a.params.size}}},{default:s(()=>[n(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),serviceType:s(({rowValue:e})=>[n(m(e||"internal"),1)]),addressPort:s(({rowValue:e})=>[e?(t(),r(h,{key:0,text:e},null,8,["text"])):(t(),_(d,{key:1},[n(m(c("common.collection.none")),1)],64))]),online:s(({row:e})=>[e.dataplanes?(t(),_(d,{key:0},[n(m(e.dataplanes.online||0)+" / "+m(e.dataplanes.total||0),1)],64)):(t(),_(d,{key:1},[n(m(c("common.collection.none")),1)],64))]),status:s(({row:e})=>[o(N,{status:e.status||"not_available"},null,8,["status"])]),details:s(({row:e})=>[o(u,{class:"details-link","data-testid":"details-link",to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.name}}},{default:s(()=>[n(m(c("common.collection.details_link"))+" ",1),o(k(T),{display:"inline-block",decorative:"",size:k(b)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","headers","page-number","page-size","total","items","error","is-selected-row","onChange"]))]),_:2},1024),n(),a.params.service?(t(),r(S,{key:0},{default:s(e=>[o(A,{onClose:f=>a.replace({name:"service-list-view",params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size}})},{default:s(()=>[(t(),r(B(e.Component),{name:a.params.service,service:i==null?void 0:i.items.find(f=>f.name===a.params.service)},null,8,["name","service"]))]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):g("",!0)]),_:1})}}});const M=D(E,[["__scopeId","data-v-0ce3984a"]]);export{M as default};
