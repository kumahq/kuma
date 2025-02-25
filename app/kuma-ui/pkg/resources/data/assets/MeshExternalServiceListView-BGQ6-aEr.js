import{d as $,r as t,o as i,m as c,w as s,b as a,e as r,p as w,as as E,q as p,aq as I,C as K,t as u,c as y,F as k,K as P}from"./index-Cm0u77zo.js";import{S as T}from"./SummaryView-2SryJn68.js";const j=$({__name:"MeshExternalServiceListView",props:{mesh:{}},setup(C){const h=C;return(F,m)=>{const b=t("RouteTitle"),v=t("XI18n"),f=t("XNotification"),d=t("XAction"),x=t("XCopyButton"),X=t("KumaPort"),V=t("XActionGroup"),A=t("RouterView"),D=t("DataCollection"),N=t("DataLoader"),R=t("XCard"),S=t("AppView"),B=t("DataSource"),q=t("RouteView");return i(),c(q,{name:"mesh-external-service-list-view",params:{page:1,size:Number,mesh:"",service:""}},{default:s(({route:n,t:_,can:L,uri:g,me:l})=>[a(b,{render:!1,title:_("services.routes.mesh-external-service-list-view.title")},null,8,["title"]),m[7]||(m[7]=r()),a(B,{src:g(w(E),"/zone-cps/:name/egresses",{name:"*"},{page:1,size:100})},{default:s(({data:z})=>[a(S,{docs:_("services.mesh-external-service.href.docs"),notifications:!0},{default:s(()=>[h.mesh.mtlsBackend?p("",!0):(i(),c(f,{key:0,uri:`mes-mtls-warning.${h.mesh.id}`},{default:s(()=>[a(v,{path:"services.mesh-external-service.notifications.mtls-warning"})]),_:1},8,["uri"])),m[5]||(m[5]=r()),z&&!z.items.find(o=>typeof o.zoneEgressInsight.connectedSubscription<"u")?(i(),c(f,{key:1,uri:`mes-no-zone-ingress.${h.mesh.id}`},{default:s(()=>[a(v,{path:"services.mesh-external-service.notifications.no-zone-egress"})]),_:1},8,["uri"])):p("",!0),m[6]||(m[6]=r()),a(R,null,{default:s(()=>[a(N,{src:g(w(I),"/meshes/:mesh/mesh-external-services",{mesh:n.params.mesh},{page:n.params.page,size:n.params.size})},{loadable:s(({data:o})=>[a(D,{type:"services",items:(o==null?void 0:o.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:o==null?void 0:o.total,onChange:n.update},{default:s(()=>[a(K,{"data-testid":"service-collection",headers:[{...l.get("headers.name"),label:"Name",key:"name"},{...l.get("headers.namespace"),label:"Namespace",key:"namespace"},...L("use zones")?[{...l.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...l.get("headers.port"),label:"Port",key:"port"},{...l.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,"is-selected-row":e=>e.name===n.params.service,onResize:l.set},{name:s(({row:e})=>[a(x,{text:e.name},{default:s(()=>[a(d,{"data-action":"",to:{name:"mesh-external-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:n.params.page,size:n.params.size}}},{default:s(()=>[r(u(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[r(u(e.namespace),1)]),zone:s(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(i(),y(k,{key:0},[e.labels["kuma.io/zone"]?(i(),c(d,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:s(()=>[r(u(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):p("",!0)],64)):(i(),y(k,{key:1},[r(u(_("common.detail.none")),1)],64))]),port:s(({row:e})=>[e.spec.match?(i(),c(X,{key:0,port:e.spec.match},null,8,["port"])):p("",!0)]),actions:s(({row:e})=>[a(V,null,{default:s(()=>[a(d,{to:{name:"mesh-external-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[r(u(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),m[4]||(m[4]=r()),o!=null&&o.items&&n.params.service?(i(),c(A,{key:0},{default:s(e=>[a(T,{onClose:G=>n.replace({name:"mesh-external-service-list-view",params:{mesh:n.params.mesh},query:{page:n.params.page,size:n.params.size}})},{default:s(()=>[(i(),c(P(e.Component),{items:o==null?void 0:o.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):p("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:2},1032,["src"])]),_:1})}}});export{j as default};
