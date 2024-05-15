import{d as w,a as n,o as i,b as r,w as e,e as a,m as z,f as c,a0 as C,T as x,t as p,c as V,F as S,p as v,Q as T,K as b,q as B,_ as E}from"./index-7WEPlSeK.js";import{A as N}from"./AppCollection-CPac3Vbi.js";const R=w({__name:"ExternalServiceListView",setup(A){return(L,D)=>{const g=n("RouteTitle"),k=n("XTeleportTemplate"),d=n("RouterLink"),y=n("KCard"),f=n("AppView"),_=n("DataSource"),h=n("RouteView");return i(),r(_,{src:"/me"},{default:e(({data:u})=>[u?(i(),r(h,{key:0,name:"external-service-list-view",params:{page:1,size:u.pageSize,mesh:""}},{default:e(({route:s,t:o})=>[a(_,{src:`/meshes/${s.params.mesh}/external-services?page=${s.params.page}&size=${s.params.size}`},{default:e(({data:l,error:m})=>[a(f,null,{title:e(()=>[a(k,{to:{name:"service-list-tabs-view-title"}},{default:e(()=>[z("h2",null,[a(g,{title:o("external-services.routes.items.title")},null,8,["title"])])]),_:2},1024)]),default:e(()=>[c(),a(y,null,{default:e(()=>[m!==void 0?(i(),r(C,{key:0,error:m},null,8,["error"])):(i(),r(N,{key:1,class:"external-service-collection","data-testid":"external-service-collection","empty-state-message":o("common.emptyState.message",{type:"External Services"}),"empty-state-cta-to":o("external-services.href.docs"),"empty-state-cta-text":o("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Address",key:"address"},{label:"Details",key:"details",hideLabel:!0}],"page-number":s.params.page,"page-size":s.params.size,total:l==null?void 0:l.total,items:l==null?void 0:l.items,error:m,onChange:s.update},{name:e(({row:t})=>[a(x,{text:t.name},{default:e(()=>[a(d,{to:{name:"external-service-detail-view",params:{mesh:t.mesh,service:t.name},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[c(p(t.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),address:e(({row:t})=>[t.networking.address?(i(),r(x,{key:0,text:t.networking.address},null,8,["text"])):(i(),V(S,{key:1},[c(p(o("common.collection.none")),1)],64))]),details:e(({row:t})=>[a(d,{class:"details-link","data-testid":"details-link",to:{name:"external-service-detail-view",params:{mesh:t.mesh,service:t.name}}},{default:e(()=>[c(p(o("common.collection.details_link"))+" ",1),a(v(T),{decorative:"",size:v(b)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):B("",!0)]),_:1})}}}),$=E(R,[["__scopeId","data-v-e24d473e"]]);export{$ as default};
