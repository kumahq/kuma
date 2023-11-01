import{d as x,g as V,e as B,r as d,o as s,l as P,j as r,w as a,F as $,I as L,B as N,k as l,n as i,H as m,p as y,i as n,m as f,E as I,a0 as E,K,t as A,x as q,q as F}from"./index-bc0f9b6f.js";import{D,A as O}from"./AppCollection-8dbcef26.js";import{P as j}from"./PolicyTypeTag-13035da4.js";import{_ as H}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-2d7e8750.js";import{S as M}from"./SummaryView-d9c36588.js";const U={class:"policy-list-content"},Z={class:"policy-count"},G={class:"policy-list"},J={class:"stack"},Q={class:"description"},W={class:"description-content"},X={class:"description-actions"},Y={class:"visually-hidden"},ee={key:0},te=x({__name:"PolicyList",props:{pageNumber:{},pageSize:{},policyTypes:{},currentPolicyType:{},policyCollection:{},policyError:{},meshInsight:{},isSelectedRow:{type:[Function,null],default:null}},emits:["change"],setup(R,{emit:T}){const{t:p}=V(),b=B(),e=R,_=T;return(S,v)=>{const h=d("RouterLink"),t=d("KCard"),g=d("KBadge");return s(),P("div",U,[r(t,{class:"policy-type-list","data-testid":"policy-type-list"},{body:a(()=>[(s(!0),P($,null,L(e.policyTypes,(c,u)=>{var o,w,C;return s(),P("div",{key:u,class:N(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":c.path===e.currentPolicyType.path}])},[r(h,{class:"policy-type-link",to:{name:"policy-list-view",params:{mesh:l(b).params.mesh,policyPath:c.path}},"data-testid":`policy-type-link-${c.name}`},{default:a(()=>[i(m(c.name),1)]),_:2},1032,["to","data-testid"]),i(),y("div",Z,m(((C=(w=(o=e.meshInsight)==null?void 0:o.policies)==null?void 0:w[c.name])==null?void 0:C.total)??0),1)],2)}),128))]),_:1}),i(),y("div",G,[y("div",J,[r(t,null,{body:a(()=>[y("div",Q,[y("div",W,[y("h3",null,[r(j,{"policy-type":e.currentPolicyType.name},{default:a(()=>[i(m(l(p)("policies.collection.title",{name:e.currentPolicyType.name})),1)]),_:1},8,["policy-type"])]),i(),y("p",null,m(l(p)(`policies.type.${e.currentPolicyType.name}.description`,void 0,{defaultMessage:l(p)("policies.collection.description")})),1)]),i(),y("div",X,[e.currentPolicyType.isExperimental?(s(),n(g,{key:0,appearance:"warning"},{default:a(()=>[i(m(l(p)("policies.collection.beta")),1)]),_:1})):f("",!0),i(),e.currentPolicyType.isInbound?(s(),n(g,{key:1,appearance:"neutral"},{default:a(()=>[i(m(l(p)("policies.collection.inbound")),1)]),_:1})):f("",!0),i(),e.currentPolicyType.isOutbound?(s(),n(g,{key:2,appearance:"neutral"},{default:a(()=>[i(m(l(p)("policies.collection.outbound")),1)]),_:1})):f("",!0),i(),r(D,{href:l(p)("policies.href.docs",{name:e.currentPolicyType.name}),"data-testid":"policy-documentation-link"},{default:a(()=>[y("span",Y,m(l(p)("common.documentation")),1)]),_:1},8,["href"])])])]),_:1}),i(),r(t,null,{body:a(()=>{var c,u;return[e.policyError!==void 0?(s(),n(I,{key:0,error:e.policyError},null,8,["error"])):(s(),n(O,{key:1,class:"policy-collection","data-testid":"policy-collection","empty-state-message":l(p)("common.emptyState.message",{type:`${e.currentPolicyType.name} policies`}),"empty-state-cta-to":l(p)("policies.href.docs",{name:e.currentPolicyType.name}),"empty-state-cta-text":l(p)("common.documentation"),headers:[{label:"Name",key:"name"},...e.currentPolicyType.isTargetRefBased?[{label:"Target ref",key:"targetRef"}]:[],{label:"Details",key:"details",hideLabel:!0}],"page-number":e.pageNumber,"page-size":e.pageSize,total:(c=e.policyCollection)==null?void 0:c.total,items:(u=e.policyCollection)==null?void 0:u.items,error:e.policyError,"is-selected-row":e.isSelectedRow,onChange:v[0]||(v[0]=o=>_("change",o))},{name:a(({rowValue:o})=>[r(h,{to:{name:"policy-summary-view",params:{mesh:l(b).params.mesh,policyPath:e.currentPolicyType.path,policy:o},query:{page:e.pageNumber,size:e.pageSize}}},{default:a(()=>[i(m(o),1)]),_:2},1032,["to"])]),targetRef:a(({row:o})=>[e.currentPolicyType.isTargetRefBased?(s(),n(g,{key:0,appearance:"neutral"},{default:a(()=>[i(m(o.spec.targetRef.kind),1),o.spec.targetRef.name?(s(),P("span",ee,[i(":"),y("b",null,m(o.spec.targetRef.name),1)])):f("",!0)]),_:2},1024)):(s(),P($,{key:1},[i(m(l(p)("common.detail.none")),1)],64))]),details:a(({row:o})=>[r(h,{class:"details-link","data-testid":"details-link",to:{name:"policy-detail-view",params:{mesh:o.mesh,policyPath:e.currentPolicyType.path,policy:o.name}}},{default:a(()=>[i(m(l(p)("common.collection.details_link"))+" ",1),r(l(E),{display:"inline-block",decorative:"",size:l(K)},null,8,["size"])]),_:2},1032,["to"])]),_:1},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error","is-selected-row"]))]}),_:1})])])])}}});const ae=A(te,[["__scopeId","data-v-949a9abb"]]),ne=x({__name:"PolicyListView",setup(R){return(T,p)=>{const b=d("RouteTitle"),e=d("RouterView"),_=d("DataSource"),S=d("AppView"),v=d("RouteView");return s(),n(_,{src:"/me"},{default:a(({data:h})=>[h?(s(),n(v,{key:0,name:"policy-list-view",params:{page:1,size:h.pageSize,mesh:"",policyPath:"",policy:""}},{default:a(({route:t,t:g})=>[r(S,null,{title:a(()=>[y("h2",null,[r(b,{title:g("policies.routes.items.title"),render:!0},null,8,["title"])])]),default:a(()=>[i(),r(_,{src:"/*/policy-types"},{default:a(({data:c,error:u})=>[u?(s(),n(I,{key:0,error:u},null,8,["error"])):c===void 0?(s(),n(q,{key:1})):c.policies.length===0?(s(),n(H,{key:2})):(s(),n(_,{key:3,src:`/meshes/${t.params.mesh}/policy-path/${t.params.policyPath}?page=${t.params.page}&size=${t.params.size}`},{default:a(({data:o,error:w})=>[r(_,{src:`/mesh-insights/${t.params.mesh}`},{default:a(({data:C})=>[(s(),n(ae,{key:t.params.policyPath,"page-number":parseInt(t.params.page),"page-size":parseInt(t.params.size),"current-policy-type":c.policies.find(k=>k.path===t.params.policyPath)??c.policies[0],"policy-types":c.policies,"mesh-insight":C,"policy-collection":o,"policy-error":w,"is-selected-row":k=>k.name===t.params.policy,onChange:t.update},null,8,["page-number","page-size","current-policy-type","policy-types","mesh-insight","policy-collection","policy-error","is-selected-row","onChange"])),i(),t.params.policy?(s(),n(e,{key:0},{default:a(k=>[r(M,{onClose:z=>t.replace({name:"policy-list-view",params:{mesh:t.params.mesh,policyPath:t.params.policyPath},query:{page:t.params.page,size:t.params.size}})},{default:a(()=>[(s(),n(F(k.Component),{name:t.params.policy,policy:o==null?void 0:o.items.find(z=>z.name===t.params.policy),"policy-type":c.policies.find(z=>z.path===t.params.policyPath)},null,8,["name","policy","policy-type"]))]),_:2},1032,["onClose"])]),_:2},1024)):f("",!0)]),_:2},1032,["src"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):f("",!0)]),_:1})}}});export{ne as default};
