import{d as $,r as t,o as l,q as f,w as o,b as i,e as u,m,c as _,K as y,L as w,n as b,t as C,G as q,_ as z}from"./index-TQmXQMPR.js";const E={class:"policy-list-content"},F={class:"policy-count"},G={class:"policy-list"},K=$({__name:"PolicyTypeListView",setup(j){return(H,c)=>{const D=t("RouteTitle"),R=t("XAction"),B=t("DataCollection"),L=t("DataLoader"),P=t("XCard"),x=t("RouterView"),v=t("DataSource"),A=t("AppView"),X=t("RouteView");return l(),f(X,{name:"policy-list-view",params:{mesh:"",policyPath:"",policy:""}},{default:o(({route:d,t:N})=>[i(D,{render:!1,title:N("policies.routes.types.title")},null,8,["title"]),c[2]||(c[2]=u()),i(A,null,{default:o(()=>[i(v,{src:`/mesh-insights/${d.params.mesh}`},{default:o(({data:e})=>[i(v,{src:"/policy-types"},{default:o(({data:s,error:S})=>[m("div",E,[i(P,{class:"policy-type-list","data-testid":"policy-type-list"},{default:o(()=>[i(L,{data:[s],errors:[S]},{default:o(()=>[(l(!0),_(y,null,w([typeof(e==null?void 0:e.policies)>"u"?s.policies:s.policies.filter(p=>{var n,a;return!p.isTargetRefBased&&(((a=(n=e.policies)==null?void 0:n[p.name])==null?void 0:a.total)??0)>0})],p=>(l(),f(B,{key:p,predicate:typeof(e==null?void 0:e.policies)>"u"?void 0:n=>p.length>0||n.isTargetRefBased,items:s.policies},{default:o(({items:n})=>[(l(!0),_(y,null,w([n.find(a=>a.path===d.params.policyPath)],a=>(l(),_(y,{key:a},[(l(!0),_(y,null,w(n,(r,T)=>{var V,k;return l(),_("div",{key:r.path,class:b(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":a&&a.path===r.path}])},[i(R,{class:"policy-type-link",to:{name:"policy-list-view",params:{mesh:d.params.mesh,policyPath:r.path}},mount:d.params.policyPath.length===0&&T===0?d.replace:void 0,"data-testid":`policy-type-link-${r.name}`},{default:o(()=>[u(C(r.name),1)]),_:2},1032,["to","mount","data-testid"]),c[0]||(c[0]=u()),m("div",F,C(((k=(V=e==null?void 0:e.policies)==null?void 0:V[r.name])==null?void 0:k.total)??0),1)],2)}),128))],64))),128))]),_:2},1032,["predicate","items"]))),128))]),_:2},1032,["data","errors"])]),_:2},1024),c[1]||(c[1]=u()),m("div",G,[i(x,null,{default:o(({Component:p})=>[(l(),f(q(p),{"policy-types":s==null?void 0:s.policies},null,8,["policy-types"]))]),_:2},1024)])])]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}}),M=z(K,[["__scopeId","data-v-5a875ff4"]]);export{M as default};
