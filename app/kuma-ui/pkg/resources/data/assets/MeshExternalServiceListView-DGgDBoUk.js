import{d as B,r as t,o as l,m as p,w as s,b as n,e as i,l as K,aB as L,A as T,U as N,t as m,c as u,F as _,p as g,s as w,E as S}from"./index-DOjcqG3h.js";import{S as E}from"./SummaryView-7OnKmF0m.js";const M=B({__name:"MeshExternalServiceListView",setup(X){return(q,F)=>{const f=t("RouteTitle"),d=t("XAction"),v=t("KBadge"),z=t("KTruncate"),k=t("XActionGroup"),b=t("RouterView"),y=t("DataCollection"),C=t("DataLoader"),x=t("KCard"),V=t("AppView"),A=t("RouteView");return l(),p(A,{name:"mesh-external-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:s(({route:o,t:h,can:R,uri:D,me:c})=>[n(f,{render:!1,title:h("services.routes.mesh-external-service-list-view.title")},null,8,["title"]),i(),n(V,null,{default:s(()=>[n(x,null,{default:s(()=>[n(C,{src:D(K(L),"/meshes/:mesh/mesh-external-services",{mesh:o.params.mesh},{page:o.params.page,size:o.params.size})},{loadable:s(({data:a})=>[n(y,{type:"services",items:(a==null?void 0:a.items)??[void 0]},{default:s(()=>[n(T,{"data-testid":"service-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.namespace"),label:"Namespace",key:"namespace"},...R("use zones")?[{...c.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...c.get("headers.tls"),label:"TLS",key:"tls"},{...c.get("headers.addresses"),label:"Addresses",key:"addresses"},{...c.get("headers.port"),label:"Port",key:"port"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,"is-selected-row":e=>e.name===o.params.service,onChange:o.update,onResize:c.set},{name:s(({row:e})=>[n(N,{text:e.name},{default:s(()=>[n(d,{"data-action":"",to:{name:"mesh-external-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:o.params.page,size:o.params.size}}},{default:s(()=>[i(m(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[i(m(e.namespace),1)]),zone:s(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(l(),u(_,{key:0},[e.labels["kuma.io/zone"]?(l(),p(d,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:s(()=>[i(m(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):g("",!0)],64)):(l(),u(_,{key:1},[i(m(h("common.detail.none")),1)],64))]),tls:s(({row:e})=>[n(v,{appearance:"neutral"},{default:s(()=>{var r;return[i(m((r=e.spec.tls)!=null&&r.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),addresses:s(({row:e})=>[n(z,null,{default:s(()=>[(l(!0),u(_,null,w(e.status.addresses,r=>(l(),u("span",{key:r.hostname},m(r.hostname),1))),128))]),_:2},1024)]),port:s(({row:e})=>[e.spec.match?(l(!0),u(_,{key:0},w([e.spec.match],r=>(l(),p(v,{key:r.port,appearance:"info"},{default:s(()=>[i(m(r.port)+"/"+m(r.protocol),1)]),_:2},1024))),128)):g("",!0)]),actions:s(({row:e})=>[n(k,null,{default:s(()=>[n(d,{to:{name:"mesh-external-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[i(m(h("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"]),i(),a!=null&&a.items&&o.params.service?(l(),p(b,{key:0},{default:s(e=>[n(E,{onClose:r=>o.replace({name:"mesh-external-service-list-view",params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size}})},{default:s(()=>[(l(),p(S(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{M as default};
