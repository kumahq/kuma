import{d as M,r as m,o as p,m as d,w as e,b as t,ac as v,e as s,t as l,k as r,p as u,l as N,aG as X,c as y,a0 as H,A as I,F as C,E as K,q as S}from"./index-Is4zmHdk.js";import{P as q}from"./PolicyTypeTag-BaBHTb6_.js";import{S as E}from"./SummaryView-QbTr0JVE.js";const F={class:"stack"},G={class:"visually-hidden"},$=["innerHTML"],O={key:0},Z=["innerHTML"],j={key:0},J=M({__name:"PolicyListView",props:{policyTypes:{}},setup(R){const w=R;return(Q,k)=>{const f=m("KBadge"),h=m("XAction"),b=m("KCard"),V=m("XInput"),T=m("XActionGroup"),z=m("DataCollection"),P=m("RouterView"),x=m("DataLoader"),A=m("AppView"),B=m("RouteView");return p(),d(B,{name:"policy-list-view",params:{page:1,size:50,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:n,t:o,can:L,uri:D,me:_})=>[t(z,{predicate:i=>typeof i<"u"&&i.path===n.params.policyPath,items:w.policyTypes??[]},{empty:e(()=>[t(v,null,{message:e(()=>[s(l(o("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:i})=>[t(A,null,{default:e(()=>[r("div",F,[t(b,null,{default:e(()=>[r("header",null,[r("div",null,[i.isExperimental?(p(),d(f,{key:0,appearance:"warning"},{default:e(()=>[s(l(o("policies.collection.beta")),1)]),_:2},1024)):u("",!0),s(),i.isInbound?(p(),d(f,{key:1,appearance:"neutral"},{default:e(()=>[s(l(o("policies.collection.inbound")),1)]),_:2},1024)):u("",!0),s(),i.isOutbound?(p(),d(f,{key:2,appearance:"neutral"},{default:e(()=>[s(l(o("policies.collection.outbound")),1)]),_:2},1024)):u("",!0),s(),t(h,{action:"docs",href:o("policies.href.docs",{name:i.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[r("span",G,l(o("common.documentation")),1)]),_:2},1032,["href"])]),s(),r("h3",null,[t(q,{"policy-type":i.name},{default:e(()=>[s(l(o("policies.collection.title",{name:i.name})),1)]),_:2},1032,["policy-type"])])]),s(),r("div",{innerHTML:o(`policies.type.${i.name}.description`,void 0,{defaultMessage:o("policies.collection.description")})},null,8,$)]),_:2},1024),s(),t(b,null,{default:e(()=>[t(x,{src:D(N(X),"/meshes/:mesh/policy-path/:path",{mesh:n.params.mesh,path:n.params.policyPath},{page:n.params.page,size:n.params.size,search:n.params.s})},{loadable:e(({data:c})=>[((c==null?void 0:c.items)??{length:0}).length>0||n.params.s.length>0?(p(),y("search",O,[r("form",{onSubmit:k[0]||(k[0]=H(()=>{},["prevent"]))},[t(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:n.params.s,debounce:1e3,onChange:a=>n.update({s:a})},null,8,["value","onChange"])],32)])):u("",!0),s(),t(z,{items:(c==null?void 0:c.items)??[void 0]},{empty:e(()=>[t(v,null,{title:e(()=>[s(l(o("policies.x-empty-state.title")),1)]),action:e(()=>[t(h,{action:"docs",href:o("policies.href.docs",{name:i.name})},{default:e(()=>[s(l(o("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[s(),r("div",{innerHTML:o("policies.x-empty-state.body",{type:i.name,suffix:n.params.s.length>0?o("common.matchingsearch"):""})},null,8,Z),s()]),_:2},1024)]),default:e(()=>[t(I,{headers:[{..._.get("headers.name"),label:"Name",key:"name"},{..._.get("headers.namespace"),label:"Namespace",key:"namespace"},...L("use zones")&&i.isTargetRefBased?[{..._.get("headers.zone"),label:"Zone",key:"zone"}]:[],...i.isTargetRefBased?[{..._.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{..._.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":n.params.page,"page-size":n.params.size,total:c==null?void 0:c.total,items:c==null?void 0:c.items,"is-selected-row":a=>a.id===n.params.policy,onChange:n.update,onResize:_.set},{name:e(({row:a})=>[t(h,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:i.path,policy:a.id},query:{page:n.params.page,size:n.params.size}}},{default:e(()=>[s(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[s(l(a.namespace.length>0?a.namespace:o("common.detail.none")),1)]),targetRef:e(({row:a})=>{var g;return[i.isTargetRefBased&&typeof((g=a.spec)==null?void 0:g.targetRef)<"u"?(p(),d(f,{key:0,appearance:"neutral"},{default:e(()=>[s(l(a.spec.targetRef.kind),1),a.spec.targetRef.name?(p(),y("span",j,[s(":"),r("b",null,l(a.spec.targetRef.name),1)])):u("",!0)]),_:2},1024)):(p(),y(C,{key:1},[s(l(o("common.detail.none")),1)],64))]}),zone:e(({row:a})=>[a.labels&&a.labels["kuma.io/origin"]==="zone"&&a.labels["kuma.io/zone"]?(p(),d(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.labels["kuma.io/zone"]}}},{default:e(()=>[s(l(a.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(p(),y(C,{key:1},[s(l(o("common.detail.none")),1)],64))]),actions:e(({row:a})=>[t(T,null,{default:e(()=>[t(h,{to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:i.path,policy:a.id}}},{default:e(()=>[s(l(o("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"])]),_:2},1032,["items"]),s(),n.params.policy?(p(),d(P,{key:1},{default:e(({Component:a})=>[t(E,{onClose:g=>n.replace({name:"policy-list-view",params:{mesh:n.params.mesh,policyPath:n.params.policyPath},query:{page:n.params.page,size:n.params.size}})},{default:e(()=>[typeof c<"u"?(p(),d(K(a),{key:0,items:c.items,"policy-type":i},null,8,["items","policy-type"])):u("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),ee=S(J,[["__scopeId","data-v-df021ead"]]);export{ee as default};
