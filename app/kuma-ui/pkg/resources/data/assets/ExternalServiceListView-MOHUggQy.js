import{d as w,i as n,o as i,a as r,w as e,j as a,g as z,k as m,a1 as C,a3 as v,t as p,b as V,H as S,A as x,K as b,e as T,_ as A}from"./index-CyAtMQ3G.js";import{p as E}from"./kong-icons.es249-vgKX97Et.js";import{A as N}from"./AppCollection-DSdmYcz_.js";import"./kong-icons.es245-BjB891cP.js";const R=w({__name:"ExternalServiceListView",setup(B){return(L,D)=>{const g=n("RouteTitle"),k=n("XTeleportTemplate"),_=n("RouterLink"),f=n("KCard"),y=n("AppView"),d=n("DataSource"),h=n("RouteView");return i(),r(d,{src:"/me"},{default:e(({data:u})=>[u?(i(),r(h,{key:0,name:"external-service-list-view",params:{page:1,size:u.pageSize,mesh:""}},{default:e(({route:s,t:o})=>[a(d,{src:`/meshes/${s.params.mesh}/external-services?page=${s.params.page}&size=${s.params.size}`},{default:e(({data:l,error:c})=>[a(y,null,{title:e(()=>[a(k,{to:{name:"service-list-tabs-view-title"}},{default:e(()=>[z("h2",null,[a(g,{title:o("external-services.routes.items.title")},null,8,["title"])])]),_:2},1024)]),default:e(()=>[m(),a(f,null,{default:e(()=>[c!==void 0?(i(),r(C,{key:0,error:c},null,8,["error"])):(i(),r(N,{key:1,class:"external-service-collection","data-testid":"external-service-collection","empty-state-message":o("common.emptyState.message",{type:"External Services"}),"empty-state-cta-to":o("external-services.href.docs"),"empty-state-cta-text":o("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"address"},{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:l==null?void 0:l.total,items:l==null?void 0:l.items,error:c,onChange:s.update},{name:e(({row:t})=>[a(v,{text:t.name},{default:e(()=>[a(_,{to:{name:"external-service-detail-view",params:{mesh:t.mesh,service:t.name},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[m(p(t.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),address:e(({row:t})=>[t.networking.address?(i(),r(v,{key:0,text:t.networking.address},null,8,["text"])):(i(),V(S,{key:1},[m(p(o("common.collection.none")),1)],64))]),details:e(({row:t})=>[a(_,{class:"details-link","data-testid":"details-link",to:{name:"external-service-detail-view",params:{mesh:t.mesh,service:t.name}}},{default:e(()=>[m(p(o("common.collection.details_link"))+" ",1),a(x(E),{decorative:"",size:x(b)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):T("",!0)]),_:1})}}}),j=A(R,[["__scopeId","data-v-e24d473e"]]);export{j as default};
