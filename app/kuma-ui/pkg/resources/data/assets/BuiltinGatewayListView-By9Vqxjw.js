import{d as b,a as l,o,b as m,w as a,e as s,E as f,T as h,f as p,t as c,c as C,F as x,p as d,Q as V,K as v,q as B,_ as L}from"./index-CvRMgvyl.js";import{A as N}from"./AppCollection-RVh7mzoi.js";const S=b({__name:"BuiltinGatewayListView",setup(A){return(D,E)=>{const r=l("RouterLink"),y=l("KCard"),g=l("AppView"),_=l("DataSource"),k=l("RouteView");return o(),m(_,{src:"/me"},{default:a(({data:w})=>[w?(o(),m(k,{key:0,name:"builtin-gateway-list-view",params:{page:1,size:10,mesh:"",gateway:""}},{default:a(({route:t,t:i,can:z})=>[s(_,{src:`/meshes/${t.params.mesh}/mesh-gateways?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:n,error:u})=>[s(g,null,{default:a(()=>[s(y,null,{default:a(()=>[u!==void 0?(o(),m(f,{key:0,error:u},null,8,["error"])):(o(),m(N,{key:1,class:"builtin-gateway-collection","data-testid":"builtin-gateway-collection","empty-state-message":i("common.emptyState.message",{type:"Built-in Gateways"}),"empty-state-cta-to":i("builtin-gateways.href.docs"),"empty-state-cta-text":i("common.documentation"),headers:[{label:"Name",key:"name"},...z("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Details",key:"details",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:n==null?void 0:n.total,items:n==null?void 0:n.items,error:u,onChange:t.update},{name:a(({row:e})=>[s(h,{text:e.name},{default:a(()=>[s(r,{to:{name:"builtin-gateway-detail-view",params:{mesh:e.mesh,gateway:e.name},query:{page:t.params.page,size:t.params.size}}},{default:a(()=>[p(c(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),zone:a(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(o(),m(r,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:a(()=>[p(c(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(o(),C(x,{key:1},[p(c(i("common.detail.none")),1)],64))]),details:a(({row:e})=>[s(r,{class:"details-link","data-testid":"details-link",to:{name:"builtin-gateway-detail-view",params:{mesh:e.mesh,gateway:e.name}}},{default:a(()=>[p(c(i("common.collection.details_link"))+" ",1),s(d(V),{decorative:"",size:d(v)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1})):B("",!0)]),_:1})}}}),R=L(S,[["__scopeId","data-v-4aae4c53"]]);export{R as default};
