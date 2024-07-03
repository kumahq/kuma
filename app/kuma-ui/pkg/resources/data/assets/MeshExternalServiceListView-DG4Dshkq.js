import{d as A,h as n,o as l,a as c,w as s,j as o,g as K,k as i,z as h,aB as N,C as E,X,t as m,b as p,F as _,e as u,B as f,M as I,r as M,A as q}from"./index-ChMk9xbI.js";import{e as F}from"./kong-icons.es249-BqzxpUAd.js";import{S as Z}from"./SummaryView-awjXeDWd.js";const $=A({__name:"MeshExternalServiceListView",setup(j){return(O,P)=>{const z=n("RouteTitle"),b=n("XTeleportTemplate"),v=n("XAction"),y=n("RouterLink"),k=n("KBadge"),g=n("KTruncate"),C=n("RouterView"),x=n("DataCollection"),V=n("DataLoader"),D=n("KCard"),S=n("AppView"),T=n("RouteView"),L=n("DataSource");return l(),c(L,{src:"/me"},{default:s(({data:w})=>[w?(l(),c(T,{key:0,name:"mesh-external-service-list-view",params:{page:1,size:w.pageSize,mesh:"",service:""}},{default:s(({route:t,t:d,can:B,uri:R})=>[o(S,null,{title:s(()=>[o(b,{to:{name:"service-list-tabs-view-title"}},{default:s(()=>[K("h2",null,[o(z,{title:d("services.routes.mesh-external-service-list-view.title")},null,8,["title"])])]),_:2},1024)]),default:s(()=>[i(),o(D,null,{default:s(()=>[o(V,{src:R(h(N),"/meshes/:mesh/mesh-external-services",{mesh:t.params.mesh},{page:t.params.page,size:t.params.size})},{loadable:s(({data:a})=>[o(x,{type:"services",items:(a==null?void 0:a.items)??[void 0]},{default:s(()=>[o(E,{"data-testid":"service-collection",headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...B("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"TLS",key:"tls"},{label:"Addresses",key:"addresses"},{label:"Port",key:"port"},{label:"Details",key:"details",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,"is-selected-row":e=>e.name===t.params.service,onChange:t.update},{name:s(({row:e})=>[o(X,{text:e.name},{default:s(()=>[o(v,{"data-action":"",to:{name:"mesh-external-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:t.params.page,size:t.params.size}}},{default:s(()=>[i(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[i(m(e.namespace),1)]),zone:s(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(l(),p(_,{key:0},[e.labels["kuma.io/zone"]?(l(),c(y,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:s(()=>[i(m(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):u("",!0)],64)):(l(),p(_,{key:1},[i(m(d("common.detail.none")),1)],64))]),tls:s(({row:e})=>[o(k,{appearance:"neutral"},{default:s(()=>{var r;return[i(m((r=e.spec.tls)!=null&&r.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),addresses:s(({row:e})=>[o(g,null,{default:s(()=>[(l(!0),p(_,null,f(e.status.addresses,r=>(l(),p("span",{key:r.hostname},m(r.hostname),1))),128))]),_:2},1024)]),port:s(({row:e})=>[e.spec.match?(l(!0),p(_,{key:0},f([e.spec.match],r=>(l(),c(k,{key:r.port,appearance:"info"},{default:s(()=>[i(m(r.port)+"/"+m(r.protocol),1)]),_:2},1024))),128)):u("",!0)]),details:s(({row:e})=>[o(v,{class:"details-link","data-testid":"details-link",to:{name:"mesh-external-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[i(m(d("common.collection.details_link"))+" ",1),o(h(F),{decorative:"",size:h(I)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange"]),i(),a!=null&&a.items&&t.params.service?(l(),c(C,{key:0},{default:s(e=>[o(Z,{onClose:r=>t.replace({name:"mesh-external-service-list-view",params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:s(()=>[(l(),c(M(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):u("",!0)]),_:1})}}}),J=q($,[["__scopeId","data-v-5790d5a6"]]);export{J as default};
