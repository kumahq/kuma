import{d as v,r as n,o as i,p,w as a,b as t,m as A,an as X,A as x,e as l,t as c,c as B,J as D,F as R,q as g}from"./index-BIN9nSPF.js";import{S as L}from"./SummaryView-CA02nKhw.js";const E=v({__name:"BuiltinGatewayListView",setup(N){return(q,_)=>{const r=n("XAction"),y=n("XCopyButton"),d=n("XActionGroup"),w=n("DataCollection"),h=n("RouterView"),z=n("DataLoader"),f=n("XCard"),b=n("AppView"),C=n("RouteView");return i(),p(C,{name:"builtin-gateway-list-view",params:{page:1,size:50,mesh:"",gateway:""}},{default:a(({route:s,t:u,can:k,me:m,uri:V})=>[t(b,{docs:u("builtin-gateways.href.docs")},{default:a(()=>[t(f,null,{default:a(()=>[t(z,{src:V(A(X),"/meshes/:mesh/mesh-gateways",{mesh:s.params.mesh},{page:s.params.page,size:s.params.size})},{loadable:a(({data:o})=>[t(w,{type:"gateways",items:(o==null?void 0:o.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:o==null?void 0:o.total,onChange:s.update},{default:a(()=>[t(x,{class:"builtin-gateway-collection","data-testid":"builtin-gateway-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},...k("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,onResize:m.set},{name:a(({row:e})=>[t(y,{text:e.name},{default:a(()=>[t(r,{"data-action":"",to:{name:"builtin-gateway-summary-view",query:{size:s.params.size,page:s.params.page},params:{mesh:e.mesh,gateway:e.id}}},{default:a(()=>[l(c(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),zone:a(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(i(),p(r,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:a(()=>[l(c(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(i(),B(D,{key:1},[l(c(u("common.detail.none")),1)],64))]),actions:a(({row:e})=>[t(d,null,{default:a(()=>[t(r,{to:{name:"builtin-gateway-detail-view",params:{mesh:e.mesh,gateway:e.name}}},{default:a(()=>[l(c(u("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"]),_[2]||(_[2]=l()),s.child()?(i(),p(h,{key:0},{default:a(({Component:e})=>[t(L,{onClose:G=>s.replace({name:"builtin-gateway-list-view",params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size}})},{default:a(()=>[typeof o<"u"?(i(),p(R(e),{key:0,items:o.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{E as default};
