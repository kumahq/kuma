import{d as x,g as L,e as B,r as g,o,l as h,j as c,w as a,F as R,I as N,B as S,k as t,n as s,H as r,p as m,a5 as T,i as p,am as b,m as f,E as I,ar as V,W as A,$ as K,K as F,as as O,t as j,x as D}from"./index-f09cca58.js";import{D as H,A as M}from"./AppCollection-4b4f9dc8.js";import{P as Q}from"./PolicyTypeTag-88e1fdf2.js";import{_ as U}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-bb9bf655.js";const W={class:"policy-list-content"},Z={class:"policy-count"},q={class:"policy-list"},G={class:"stack"},J={class:"description"},X={class:"description-content"},Y={class:"description-actions"},ee={class:"visually-hidden"},te={key:0},ae=x({__name:"PolicyList",props:{pageNumber:{},pageSize:{},policyTypes:{},currentPolicyType:{},policyCollection:{},policyError:{},meshInsight:{}},emits:["change"],setup(z,{emit:w}){const{t:l}=L(),u=B(),e=z,v=w;return(C,_)=>{const n=g("RouterLink");return o(),h("div",W,[c(t(T),{class:"policy-type-list","data-testid":"policy-type-list"},{body:a(()=>[(o(!0),h(R,null,N(e.policyTypes,(y,d)=>{var i,k,P;return o(),h("div",{key:d,class:S(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":y.path===e.currentPolicyType.path}])},[c(n,{class:"policy-type-link",to:{name:"policy-list-view",params:{mesh:t(u).params.mesh,policyPath:y.path}},"data-testid":`policy-type-link-${y.name}`},{default:a(()=>[s(r(y.name),1)]),_:2},1032,["to","data-testid"]),s(),m("div",Z,r(((P=(k=(i=e.meshInsight)==null?void 0:i.policies)==null?void 0:k[y.name])==null?void 0:P.total)??0),1)],2)}),128))]),_:1}),s(),m("div",q,[m("div",G,[c(t(T),null,{body:a(()=>[m("div",J,[m("div",X,[m("h3",null,[c(Q,{"policy-type":e.currentPolicyType.name},{default:a(()=>[s(r(t(l)("policies.collection.title",{name:e.currentPolicyType.name})),1)]),_:1},8,["policy-type"])]),s(),m("p",null,r(t(l)(`policies.type.${e.currentPolicyType.name}.description`,void 0,{defaultMessage:t(l)("policies.collection.description")})),1)]),s(),m("div",Y,[e.currentPolicyType.isExperimental?(o(),p(t(b),{key:0,appearance:"warning"},{default:a(()=>[s(r(t(l)("policies.collection.beta")),1)]),_:1})):f("",!0),s(),e.currentPolicyType.isInbound?(o(),p(t(b),{key:1,appearance:"neutral"},{default:a(()=>[s(r(t(l)("policies.collection.inbound")),1)]),_:1})):f("",!0),s(),e.currentPolicyType.isOutbound?(o(),p(t(b),{key:2,appearance:"neutral"},{default:a(()=>[s(r(t(l)("policies.collection.outbound")),1)]),_:1})):f("",!0),s(),c(H,{href:t(l)("policies.href.docs",{name:e.currentPolicyType.name}),"data-testid":"policy-documentation-link"},{default:a(()=>[m("span",ee,r(t(l)("common.documentation")),1)]),_:1},8,["href"])])])]),_:1}),s(),c(t(T),null,{body:a(()=>{var y,d;return[e.policyError!==void 0?(o(),p(I,{key:0,error:e.policyError},null,8,["error"])):(o(),p(M,{key:1,class:"policy-collection","data-testid":"policy-collection","empty-state-message":t(l)("common.emptyState.message",{type:`${e.currentPolicyType.name} policies`}),"empty-state-cta-to":t(l)("policies.href.docs",{name:e.currentPolicyType.name}),"empty-state-cta-text":t(l)("common.documentation"),headers:[{label:"Name",key:"name"},...e.currentPolicyType.isTargetRefBased?[{label:"Target ref",key:"targetRef"}]:[],{label:"Actions",key:"actions",hideLabel:!0}],"page-number":e.pageNumber,"page-size":e.pageSize,total:(y=e.policyCollection)==null?void 0:y.total,items:(d=e.policyCollection)==null?void 0:d.items,error:e.policyError,onChange:_[0]||(_[0]=i=>v("change",i))},{name:a(({rowValue:i})=>[c(n,{to:{name:"policy-detail-view",params:{mesh:t(u).params.mesh,policyPath:e.currentPolicyType.path,policy:i}}},{default:a(()=>[s(r(i),1)]),_:2},1032,["to"])]),targetRef:a(({row:i})=>[e.currentPolicyType.isTargetRefBased?(o(),p(t(b),{key:0,appearance:"neutral"},{default:a(()=>[s(r(i.spec.targetRef.kind),1),i.spec.targetRef.name?(o(),h("span",te,[s(":"),m("b",null,r(i.spec.targetRef.name),1)])):f("",!0)]),_:2},1024)):(o(),h(R,{key:1},[s(r(t(l)("common.detail.none")),1)],64))]),actions:a(({row:i})=>[c(t(V),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:a(()=>[c(t(A),{class:"non-visual-button",appearance:"secondary",size:"small"},{default:a(()=>[c(t(K),{size:t(F)},null,8,["size"])]),_:1})]),items:a(()=>[c(t(O),{item:{to:{name:"policy-detail-view",params:{mesh:t(u).params.mesh,policyPath:e.currentPolicyType.path,policy:i.name}},label:t(l)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:1},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"]))]}),_:1})])])])}}});const se=j(ae,[["__scopeId","data-v-9ebcab5f"]]),ne=x({__name:"PolicyListView",setup(z){return(w,l)=>{const u=g("RouteTitle"),e=g("DataSource"),v=g("AppView"),C=g("RouteView");return o(),p(e,{src:"/me"},{default:a(({data:_})=>[_?(o(),p(C,{key:0,name:"policy-list-view",params:{page:1,size:_.pageSize,mesh:"",policyPath:""}},{default:a(({route:n,t:y})=>[c(v,null,{title:a(()=>[m("h2",null,[c(u,{title:y("policies.routes.items.title"),render:!0},null,8,["title"])])]),default:a(()=>[s(),c(e,{src:"/*/policy-types"},{default:a(({data:d,error:i})=>[i?(o(),p(I,{key:0,error:i},null,8,["error"])):d===void 0?(o(),p(D,{key:1})):d.policies.length===0?(o(),p(U,{key:2})):(o(),p(e,{key:3,src:`/meshes/${n.params.mesh}/policy-path/${n.params.policyPath}?page=${n.params.page}&size=${n.params.size}`},{default:a(({data:k,error:P})=>[c(e,{src:`/mesh-insights/${n.params.mesh}`},{default:a(({data:$})=>[(o(),p(se,{key:n.params.policyPath,"page-number":parseInt(n.params.page),"page-size":parseInt(n.params.size),"current-policy-type":d.policies.find(E=>E.path===n.params.policyPath)??d.policies[0],"policy-types":d.policies,"mesh-insight":$,"policy-collection":k,"policy-error":P,onChange:n.update},null,8,["page-number","page-size","current-policy-type","policy-types","mesh-insight","policy-collection","policy-error","onChange"]))]),_:2},1032,["src"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):f("",!0)]),_:1})}}});export{ne as default};
