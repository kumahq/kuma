import{d as D,r as m,o as p,m as r,w as e,b as i,a4 as z,e as a,t as l,k as d,p as _,av as I,l as N,ax as H,A as K,c as g,L as w,E as S,q as $}from"./index-CMjLgvOo.js";import{P as q}from"./PolicyTypeTag-ELdipR84.js";import{S as E}from"./SummaryView-C546ionl.js";const F={class:"stack"},G={class:"visually-hidden"},O=["innerHTML"],Z=["innerHTML"],j={key:0},J=D({__name:"PolicyListView",props:{policyTypes:{}},setup(R){const C=R;return(Q,k)=>{const h=m("KBadge"),f=m("XAction"),b=m("KCard"),V=m("XInput"),L=m("XIcon"),P=m("XActionGroup"),v=m("DataCollection"),T=m("RouterView"),x=m("DataLoader"),A=m("AppView"),B=m("RouteView");return p(),r(B,{name:"policy-list-view",params:{page:1,size:50,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:o,t:s,can:X,uri:M,me:u})=>[i(v,{predicate:t=>typeof t<"u"&&t.path===o.params.policyPath,items:C.policyTypes??[]},{empty:e(()=>[i(z,null,{message:e(()=>[a(l(s("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:t})=>[i(A,null,{default:e(()=>[d("div",F,[i(b,null,{default:e(()=>[d("header",null,[d("div",null,[t.isExperimental?(p(),r(h,{key:0,appearance:"warning"},{default:e(()=>[a(l(s("policies.collection.beta")),1)]),_:2},1024)):_("",!0),a(),t.isInbound?(p(),r(h,{key:1,appearance:"neutral"},{default:e(()=>[a(l(s("policies.collection.inbound")),1)]),_:2},1024)):_("",!0),a(),t.isOutbound?(p(),r(h,{key:2,appearance:"neutral"},{default:e(()=>[a(l(s("policies.collection.outbound")),1)]),_:2},1024)):_("",!0),a(),i(f,{action:"docs",href:s("policies.href.docs",{name:t.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",G,l(s("common.documentation")),1)]),_:2},1032,["href"])]),a(),d("h3",null,[i(q,{"policy-type":t.name},{default:e(()=>[a(l(s("policies.collection.title",{name:t.name})),1)]),_:2},1032,["policy-type"])])]),a(),d("div",{innerHTML:s(`policies.type.${t.name}.description`,void 0,{defaultMessage:s("policies.collection.description")})},null,8,O)]),_:2},1024),a(),i(b,null,{default:e(()=>[d("search",null,[d("form",{onSubmit:k[0]||(k[0]=I(()=>{},["prevent"]))},[i(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:o.params.s,debounce:1e3,onChange:c=>o.update({s:c})},null,8,["value","onChange"])],32)]),a(),i(x,{src:M(N(H),"/meshes/:mesh/policy-path/:path",{mesh:o.params.mesh,path:o.params.policyPath},{page:o.params.page,size:o.params.size,search:o.params.s})},{loadable:e(({data:c})=>[i(v,{items:(c==null?void 0:c.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:c==null?void 0:c.total,onChange:o.update},{empty:e(()=>[i(z,null,{title:e(()=>[a(l(s("policies.x-empty-state.title")),1)]),action:e(()=>[i(f,{action:"docs",href:s("policies.href.docs",{name:t.name})},{default:e(()=>[a(l(s("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[a(),d("div",{innerHTML:s("policies.x-empty-state.body",{type:t.name,suffix:o.params.s.length>0?s("common.matchingsearch"):""})},null,8,Z),a()]),_:2},1024)]),default:e(()=>[i(K,{headers:[{...u.get("headers.role"),label:"Role",key:"role",hideLabel:!0},{...u.get("headers.name"),label:"Name",key:"name"},{...u.get("headers.namespace"),label:"Namespace",key:"namespace"},...X("use zones")&&t.isTargetRefBased?[{...u.get("headers.zone"),label:"Zone",key:"zone"}]:[],...t.isTargetRefBased?[{...u.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{...u.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:c==null?void 0:c.items,"is-selected-row":n=>n.id===o.params.policy,onResize:u.set},{role:e(({row:n})=>[n.role==="producer"?(p(),r(L,{key:0,name:`policy-role-${n.role}`},{default:e(()=>[a(`
                              Role: `+l(n.role),1)]),_:2},1032,["name"])):(p(),g(w,{key:1},[a(`
                             
                          `)],64))]),name:e(({row:n})=>[i(f,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:n.mesh,policyPath:t.path,policy:n.id},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[a(l(n.name),1)]),_:2},1032,["to"])]),namespace:e(({row:n})=>[a(l(n.namespace.length>0?n.namespace:s("common.detail.none")),1)]),targetRef:e(({row:n})=>{var y;return[typeof((y=n.spec)==null?void 0:y.targetRef)<"u"?(p(),r(h,{key:0,appearance:"neutral"},{default:e(()=>[a(l(n.spec.targetRef.kind),1),n.spec.targetRef.name?(p(),g("span",j,[a(":"),d("b",null,l(n.spec.targetRef.name),1)])):_("",!0)]),_:2},1024)):(p(),r(h,{key:1,appearance:"neutral"},{default:e(()=>[a(`
                            Mesh
                          `)]),_:1}))]}),zone:e(({row:n})=>[n.zone?(p(),r(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:e(()=>[a(l(n.zone),1)]),_:2},1032,["to"])):(p(),g(w,{key:1},[a(l(s("common.detail.none")),1)],64))]),actions:e(({row:n})=>[i(P,null,{default:e(()=>[i(f,{to:{name:"policy-detail-view",params:{mesh:n.mesh,policyPath:t.path,policy:n.id}}},{default:e(()=>[a(l(s("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"]),a(),o.params.policy?(p(),r(T,{key:0},{default:e(({Component:n})=>[i(E,{onClose:y=>o.replace({name:"policy-list-view",params:{mesh:o.params.mesh,policyPath:o.params.policyPath},query:{page:o.params.page,size:o.params.size}})},{default:e(()=>[typeof c<"u"?(p(),r(S(n),{key:0,items:c.items,"policy-type":t},null,8,["items","policy-type"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),ee=$(J,[["__scopeId","data-v-7f8fcb50"]]);export{ee as default};
