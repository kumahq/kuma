import{d as $,r as t,m as f,o as l,w as o,b as i,e as y,s as b,V as q,q as m,c as _,F as u,v as w,n as z,t as C,K as E,_ as K}from"./index-D_WxlpfD.js";const j={class:"policy-list-content"},G={class:"policy-count"},H={class:"policy-list"},J=$({__name:"PolicyTypeListView",setup(M){return(O,n)=>{const D=t("RouteTitle"),R=t("XAction"),P=t("DataCollection"),T=t("DataLoader"),x=t("XCard"),L=t("RouterView"),v=t("DataSource"),A=t("AppView"),B=t("RouteView");return l(),f(B,{name:"policy-list-view",params:{mesh:"",policyPath:"",policy:""}},{default:o(({uri:X,route:d,t:N})=>[i(D,{render:!1,title:N("policies.routes.types.title")},null,8,["title"]),n[2]||(n[2]=y()),i(A,null,{default:o(()=>[i(v,{src:`/mesh-insights/${d.params.mesh}`},{default:o(({data:e})=>[i(v,{src:X(b(q),"/policy-types",{})},{default:o(({data:p,error:S})=>[m("div",j,[i(x,{class:"policy-type-list","data-testid":"policy-type-list"},{default:o(()=>[i(T,{data:[p],errors:[S]},{default:o(()=>[(l(!0),_(u,null,w([typeof(e==null?void 0:e.policies)>"u"?p.policyTypes:p.policyTypes.filter(s=>{var c,a;return!s.policy.isTargetRef&&(((a=(c=e.policies)==null?void 0:c[s.name])==null?void 0:a.total)??0)>0})],s=>(l(),f(P,{key:s,predicate:typeof(e==null?void 0:e.policies)>"u"?void 0:c=>s.length>0||c.policy.isTargetRef,items:p.policyTypes},{default:o(({items:c})=>[(l(!0),_(u,null,w([c.find(a=>a.path===d.params.policyPath)],a=>(l(),_(u,{key:a},[(l(!0),_(u,null,w(c,(r,F)=>{var V,k;return l(),_("div",{key:r.path,class:z(["policy-type-link-wrapper",{"policy-type-link-wrapper--is-active":a&&a.path===r.path}])},[i(R,{class:"policy-type-link",to:{name:"policy-list-view",params:{mesh:d.params.mesh,policyPath:r.path}},mount:d.params.policyPath.length===0&&F===0?d.replace:void 0,"data-testid":`policy-type-link-${r.name}`},{default:o(()=>[y(C(r.name),1)]),_:2},1032,["to","mount","data-testid"]),n[0]||(n[0]=y()),m("div",G,C(((k=(V=e==null?void 0:e.policies)==null?void 0:V[r.name])==null?void 0:k.total)??0),1)],2)}),128))],64))),128))]),_:2},1032,["predicate","items"]))),128))]),_:2},1032,["data","errors"])]),_:2},1024),n[1]||(n[1]=y()),m("div",H,[i(L,null,{default:o(({Component:s})=>[(l(),f(E(s),{"policy-types":p==null?void 0:p.policyTypes},null,8,["policy-types"]))]),_:2},1024)])])]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}}),U=K(J,[["__scopeId","data-v-e5fa100c"]]);export{U as default};
