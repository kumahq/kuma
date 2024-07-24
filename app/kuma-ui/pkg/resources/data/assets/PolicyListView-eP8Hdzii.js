import{d as N,r as m,o as p,m as r,w as e,b as t,ab as v,e as s,t as l,k as d,p as u,l as X,aE as E,$ as H,A as I,c as g,F as C,E as K,q as S}from"./index--1DEc0sn.js";import{P as $}from"./PolicyTypeTag-Co3RB2NR.js";import{S as q}from"./SummaryView-D30ZTwxO.js";const F={class:"stack"},G={class:"visually-hidden"},O=["innerHTML"],Z=["innerHTML"],j={key:0},J=N({__name:"PolicyListView",props:{policyTypes:{}},setup(R){const w=R;return(Q,k)=>{const f=m("KBadge"),h=m("XAction"),b=m("KCard"),V=m("XInput"),T=m("search"),P=m("XActionGroup"),z=m("DataCollection"),x=m("RouterView"),A=m("DataLoader"),B=m("AppView"),L=m("RouteView");return p(),r(L,{name:"policy-list-view",params:{page:1,size:50,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:n,t:o,can:D,uri:M,me:_})=>[t(z,{predicate:i=>typeof i<"u"&&i.path===n.params.policyPath,items:w.policyTypes??[]},{empty:e(()=>[t(v,null,{message:e(()=>[s(l(o("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:i})=>[t(B,null,{default:e(()=>[d("div",F,[t(b,null,{default:e(()=>[d("header",null,[d("div",null,[i.isExperimental?(p(),r(f,{key:0,appearance:"warning"},{default:e(()=>[s(l(o("policies.collection.beta")),1)]),_:2},1024)):u("",!0),s(),i.isInbound?(p(),r(f,{key:1,appearance:"neutral"},{default:e(()=>[s(l(o("policies.collection.inbound")),1)]),_:2},1024)):u("",!0),s(),i.isOutbound?(p(),r(f,{key:2,appearance:"neutral"},{default:e(()=>[s(l(o("policies.collection.outbound")),1)]),_:2},1024)):u("",!0),s(),t(h,{type:"docs",href:o("policies.href.docs",{name:i.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",G,l(o("common.documentation")),1)]),_:2},1032,["href"])]),s(),d("h3",null,[t($,{"policy-type":i.name},{default:e(()=>[s(l(o("policies.collection.title",{name:i.name})),1)]),_:2},1032,["policy-type"])])]),s(),d("div",{innerHTML:o(`policies.type.${i.name}.description`,void 0,{defaultMessage:o("policies.collection.description")})},null,8,O)]),_:2},1024),s(),t(b,null,{default:e(()=>[t(A,{src:M(X(E),"/meshes/:mesh/policy-path/:path",{mesh:n.params.mesh,path:n.params.policyPath},{page:n.params.page,size:n.params.size,search:n.params.s})},{loadable:e(({data:c})=>[((c==null?void 0:c.items)??{length:0}).length>0||n.params.s.length>0?(p(),r(T,{key:0},{default:e(()=>[d("form",{onSubmit:k[0]||(k[0]=H(()=>{},["prevent"]))},[t(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:n.params.s,debounce:1e3,onChange:a=>n.update({s:a})},null,8,["value","onChange"])],32)]),_:2},1024)):u("",!0),s(),t(z,{items:(c==null?void 0:c.items)??[void 0]},{empty:e(()=>[t(v,null,{title:e(()=>[s(l(o("policies.x-empty-state.title")),1)]),action:e(()=>[t(h,{type:"docs",href:o("policies.href.docs",{name:i.name})},{default:e(()=>[s(l(o("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[s(),d("div",{innerHTML:o("policies.x-empty-state.body",{type:i.name,suffix:n.params.s.length>0?o("common.matchingsearch"):""})},null,8,Z),s()]),_:2},1024)]),default:e(()=>[t(I,{headers:[{..._.get("headers.name"),label:"Name",key:"name"},{..._.get("headers.namespace"),label:"Namespace",key:"namespace"},...D("use zones")&&i.isTargetRefBased?[{..._.get("headers.zone"),label:"Zone",key:"zone"}]:[],...i.isTargetRefBased?[{..._.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{..._.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":n.params.page,"page-size":n.params.size,total:c==null?void 0:c.total,items:c==null?void 0:c.items,"is-selected-row":a=>a.id===n.params.policy,onChange:n.update,onResize:_.set},{name:e(({row:a})=>[t(h,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:i.path,policy:a.id},query:{page:n.params.page,size:n.params.size}}},{default:e(()=>[s(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[s(l(a.namespace.length>0?a.namespace:o("common.detail.none")),1)]),targetRef:e(({row:a})=>{var y;return[i.isTargetRefBased&&typeof((y=a.spec)==null?void 0:y.targetRef)<"u"?(p(),r(f,{key:0,appearance:"neutral"},{default:e(()=>[s(l(a.spec.targetRef.kind),1),a.spec.targetRef.name?(p(),g("span",j,[s(":"),d("b",null,l(a.spec.targetRef.name),1)])):u("",!0)]),_:2},1024)):(p(),g(C,{key:1},[s(l(o("common.detail.none")),1)],64))]}),zone:e(({row:a})=>[a.labels&&a.labels["kuma.io/origin"]==="zone"&&a.labels["kuma.io/zone"]?(p(),r(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.labels["kuma.io/zone"]}}},{default:e(()=>[s(l(a.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):(p(),g(C,{key:1},[s(l(o("common.detail.none")),1)],64))]),actions:e(({row:a})=>[t(P,null,{default:e(()=>[t(h,{to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:i.path,policy:a.id}}},{default:e(()=>[s(l(o("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"])]),_:2},1032,["items"]),s(),n.params.policy?(p(),r(x,{key:1},{default:e(({Component:a})=>[t(q,{onClose:y=>n.replace({name:"policy-list-view",params:{mesh:n.params.mesh,policyPath:n.params.policyPath},query:{page:n.params.page,size:n.params.size}})},{default:e(()=>[typeof c<"u"?(p(),r(K(a),{key:0,items:c.items,"policy-type":i},null,8,["items","policy-type"])):u("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),ee=S(J,[["__scopeId","data-v-9fa67f3c"]]);export{ee as default};
