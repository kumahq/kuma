import{d as x,r as X,o as e,e as s,g as a,F as v,s as G,q as y,t as p,h as o,w as t,f as Y,a as m,B as Ce,b as f,Y as ke,p as _e,m as Oe,c as R,v as he,z as ae,$ as ve,j as N,I as Ee,K as Pe,G as H,L as Ge,M as Qe}from"./index-d1d97159.js";import{A as Z,a as $,S as Ie,b as Ue}from"./SubscriptionHeader-7e0e181f.js";import{f as S,e as se,m as ge,p as fe,E as we,q as Te,g as be,D as V,S as Le,A as ze,o as xe,_ as Me}from"./RouteView.vue_vue_type_script_setup_true_lang-3fa7796e.js";import{_ as De}from"./CodeBlock.vue_vue_type_style_index_0_lang-b366b6e9.js";import{T as j}from"./TagList-c0a0c2f1.js";import{t as ie,_ as Re}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-cee11945.js";import{_ as Ye,E as te}from"./EnvoyData-6c79dcf5.js";import{T as Se}from"./TabsWidget-2efbb2f2.js";import{T as Ne}from"./TextWithCopyButton-aee4bf2b.js";import{_ as He}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-e7b0bce0.js";import{a as qe,d as oe,b as Ke,p as je,c as Fe,C as Je,I as We,e as Ve}from"./dataplane-30467516.js";import{_ as Xe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-b054f500.js";import"./CopyButton-fc519f3f.js";const F=r=>(_e("data-v-d6898838"),r=r(),Oe(),r),Ze={class:"mesh-gateway-policy-list"},$e=F(()=>y("h3",{class:"mb-2"},`
      Gateway policies
    `,-1)),et={key:0},tt=F(()=>y("h3",{class:"mt-6 mb-2"},`
      Listeners
    `,-1)),at=F(()=>y("b",null,"Host",-1)),st=F(()=>y("h4",{class:"mt-2 mb-2"},`
              Routes
            `,-1)),nt={class:"dataplane-policy-header"},it=F(()=>y("b",null,"Route",-1)),ot=F(()=>y("b",null,"Service",-1)),lt={key:0,class:"badge-list"},At={class:"mt-1"},ct=x({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(r){const n=r;return(u,_)=>{const D=X("router-link");return e(),s("div",Ze,[$e,a(),r.meshGatewayRoutePolicies.length>0?(e(),s("ul",et,[(e(!0),s(v,null,G(r.meshGatewayRoutePolicies,(h,T)=>(e(),s("li",{key:T},[y("span",null,p(h.type),1),a(`:

        `),o(D,{to:h.route},{default:t(()=>[a(p(h.name),1)]),_:2},1032,["to"])]))),128))])):Y("",!0),a(),tt,a(),y("div",null,[(e(!0),s(v,null,G(n.meshGatewayListenerEntries,(h,T)=>(e(),s("div",{key:T},[y("div",null,[y("div",null,[at,a(": "+p(h.hostName)+":"+p(h.port)+" ("+p(h.protocol)+`)
          `,1)]),a(),h.routeEntries.length>0?(e(),s(v,{key:0},[st,a(),o($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,G(h.routeEntries,(A,b)=>(e(),m(Z,{key:b},Ce({"accordion-header":t(()=>[y("div",nt,[y("div",null,[y("div",null,[it,a(": "),o(D,{to:A.route},{default:t(()=>[a(p(A.routeName),1)]),_:2},1032,["to"])]),a(),y("div",null,[ot,a(": "+p(A.service),1)])]),a(),A.policies.length>0?(e(),s("div",lt,[(e(!0),s(v,null,G(A.policies,(i,w)=>(e(),m(f(ke),{key:`${T}-${w}`},{default:t(()=>[a(p(i.type),1)]),_:2},1024))),128))])):Y("",!0)])]),_:2},[A.policies.length>0?{name:"accordion-content",fn:t(()=>[y("ul",At,[(e(!0),s(v,null,G(A.policies,(i,w)=>(e(),s("li",{key:`${T}-${w}`},[a(p(i.type)+`:

                      `,1),o(D,{to:i.route},{default:t(()=>[a(p(i.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):Y("",!0)])]))),128))])])}}});const rt=S(ct,[["__scopeId","data-v-d6898838"]]),le="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB4CAMAAAAOusbgAAAAVFBMVEXa2tra2tra2tra2tra2tra2tr////a2toAfd6izPLvzPnRfvDYteSKr86zas0Aar4AhODY6vr3+Prx8v2Kv+9aqOk3muUOj+N5t+211vXhqfW01fXvn55GAAAABnRSTlMC9s/Hbhsvz/I3AAABVklEQVRo3u3b3Y6CMBCG4SJYhnV/KD+K7v3f57bN7AFJTcDUmZB+74lH5EmMA5hmjK+pq1awqm5M6HxqxTudPSzssmxM06rUmDp8DFawIYi1qYRdlisTeCtcMAGnAgwYMGDAgJ8GGPDB4B8frepnl9cZH5d1374E7GmX1WVuA0xzTvixA+5zwpc0/OXrVgU5N/yx6tMHGDBgwIABvxmeiBZhmF3fPMjDFLuOSjDdnBJMvVOAb1G+y8PjlUKdOGyHOcpLJniiDfEVC/FYZYA3unxFx2OVAd7sTjZ073msRGB2Yy7KvcsC2z05Hitx2P6PVTEwf9W/h/5xvTBOB76ByN8ydzRRzofELln1schjVNCrTxyjsl5vtV7ol7L+tAEGDLhMWOAw5ADHPxIHXmpHfAWepgJOBBgwYMCAAT8NMGDAgJOw2hKO2tqR2qKV1mqZ3jKd2vrgH/W3idgykdWgAAAAAElFTkSuQmCC",ut="data:image/png;base64,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",Ae="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEj0lEQVRoBe1aS28TVxT+7ngydhwnPGISTBKHEB6iUtOWHTvWqCtKqQhISC2vBbCpRDf9BUgsgAWbdlGppWqlSl1UXfMLCmXRqgXUxmCclOCWxI4Tv4Zz7s00BntmLh4rTCSfxJ4755458333fHfuTTQCZFOHTo+ijCs2cAi2nWJfaE2InABuw8Lle7e/eCwYvL2CXwF7a2hBtwQm8iKKdwwe+Y0HnhnRgBN2Q8qmJcPwOxm7EXrNe40jzVfDq38j9HUJvOkqdSvQrUDAEeiIhGaPH8bsyfe1oWQuTuPxhePa8V6BplenTl85tQ2l9A7YMUsnHMsTIyjtm9CK1QkKXIHC3nEI2l3RgqhzPzw/sB/g+A5ZYAKlPTsVFMnCH1Xx3f26XP2TUUQgAuXhJKr9fQqQRgVYPpUtA7IANvQq5sciEIHi7jHKb5OE9DQh5SOvoGs6pKNABJYn06tAaDQ1SLB82DoFnnO1TaA8NIhqIo7IQkFLDI58zPx/WvEMTsfaJlAiPbPF789oiWHxPTX6A3f/kPGdmQEBKlCaGJUE+oiANJ9JvEAEeOL23/ldHvVmjUrt9d1WBSrJLaiRfMzCEqzcU8pPcDzmAMunSk8f699FxP7KqngvVK/R19ZKvDy+Qy5cvQ8z8la2xuhzII8+m9foF9+axOz0YRm3/dbP6PvtoWy7fZm1iIV6tAd1i4+W3BLUrR7Y1Jb+1T7eKqg41ccajj94JPPy4DskaoleZM8cRYmeUGyO1hm0Q6DRz5XMnj2KpV1jTcSYyOTnNzjc1Uw1eCwBpQIFhNWqfvhKCZDPZbCQoGK5eVhz82uJKYjBPDp/DFwhBswZnEcmT3YlnzV/jRbBzKVplFNDTeDXEnu3TLNeBpb44x3o20vksh8fQYU2d1GaF+nr3yBCc6SVOaQyl05gxYm/9rWMf1VCra5v9LU1BxoT/N+mCpSHB2HNzmP05neu4J14ltZKKqnIroLnPta8n2ycHHzsHAGqgPXPM4x8+QOBLzXeo6ntSMsiGaYbwDcFajg6QiA6k0M9EQM/NSJFb/CMqe/PDD0QTKrU976V8uMg3j74ifOg8IsNZX9bC1mYmHQJvOlqBJ7EcUPgw8EELFq5vn1WQKHmPaX6IwIXhzdJ3jfmnmPRJ95vgAJJqJfAf0Tgx3pMpGn7cW5oExIE0M0Y/GepzdgT65EfbrPvVZuKW7g6vlV+uO1lYurgWTtmGHIEo7QYxYhSlM6jlJf9UT6nNvtiBFj5+SjUNeRbrNWpLTBmRSiOc6h8bjfOlquya8TyEQDdN1+t4dOZvFsqXsjU3ob/rqVfMv5iGaijbdORO2ihUlshiqdu5RZ4Uqnix3wRBsWcSiawj/8/xAEqGSd8ye4vV8DS4e3EheEBWYmXAl7zJJTrAMvm1LaEpPLV0wLu8V7NxUJJwAVrS3egSdwy4zo7uwTWecCbbtetQNOQrLPDoOd1bp3v2bnbEXZaN+nFiQ1qjJ3WfFymZdN9rQ4tOcJM2CNzf/+ysH33gVuiLlIkpyTh7Q8tZgbGr9sI8RO9qfIBv27zAiEVYZQrGIvuAAAAAElFTkSuQmCC",ce="data:image/png;base64,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",re="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAHgAAAB5CAYAAADyOOV3AAAFOklEQVR4Ae2dP2tUQRTFp7S385MofgRFiGBhKr9GuqBiI9iktwosCAnRLo0g8Q+ojSaNBomumESTIAqSLiPTTc4k7+bmztudu3sCAztv7p137/nNebtuREMIIXz9eXBluLO/NNzZe8sxCRrsL23tHlxObMP33b3ZzeHO0edv25FjcjRITBPbsPXj12+CnRywOcvENtC9kwk3gU5sQ048vf7775DDsQbIswAc+eNaAQJ2jU8unoBljVxHELBrfHLxBCxr5DqCgF3jk4snYFkj1xEE7BqfXDwByxq5jiBg1/jk4glY1sh1BAG7xicXT8CyRq4jCNg1Prl4ApY1ch1BwK7xycUTsKyR6wgCdo1PLp6AZY1cRxCwa3xy8QQsa+Q6goBd45OLJ2BZI9cRBOwan1w8AcsauY4gYNf45OIJWNbIdQQBu8YnF0/AskauIwjYNT65eAKWNXIdUQD+c2sm5iPemY2mIcnD/bsVMuqTs0yvQ7wQYtXRXb79XtxfpSEB4wH3foCgHwIGQSS+5qeddAOsxzgPOwsLMR9xsBhNQ2qA+3crZNQnZ5le89/o6Jbb3WrxKRovuOuIBR9TAHnSwcfk8T8hYP8MOzsg4E55/C8SsH+GnR0QcKc8/hcJ2D/Dzg4IuFMe/4sE7J9hZwcE3CmP/8WpAzz7cCnm48bdQaw58r3T63H/TB3gcG0+jnIQ8IgVGCXcdK9x/9DBPTuagEesADr43uBFrDlwf217+B5unV+fX4z5mPjfJiGA95vbsebA/bWAMb/6HJ/Z2gJbj0fBasJNe+H+Wj0wv/qcgG2ORiAErFXAGI8AWnfwo5U30TLmHq/GfPA92PiejAdIex4x33oAl9c+xnwQMAFrz2Rb8bUdgg7D/bXdYz7ur53n7k2v6WA6WHsm24qv7RB0FO6v7R7zcX/tnA42OhYFR0AErFXAGI8AEJB1jvtry8V8az2Fg/PvLdNr63ehmK9tuHZ8bQERAO6vrR/zcX/tvACMN6g91zZcOx770QomxeP+2voxX7qftE7A0/YejCeo9lx7omvHYz+SA7TruL+2fszX3h/jCwfn31um15bvQVMuFqxtGN/DrXOsBwWxznF/bb+Yb62nAIwXrDfAgq0N437WubU/zMd6rP3i/to58gx4QbshxtduGPezzrFe6xzrIWBBARSs9twKFPOxPqG9YhnzcX/tHA3bvIOtnwkwXyuYFI+ACoLCBcyX7ietuwMsNTTudQQk8CyWMd/aDwFP25+Dkbj1BOGJLI6scAHzrfX0nY/1Cu0Vy5hvrRd5Nv8ebG2473wEVBAULmC+tV4C5iO6rb9Gaj3RfeejAwXDFsuYb62XDqaD6WCNi9CBhUWFC5ivufdJsXQwHUwHn+SM066hAwXDFsuYf9p9znqdDqaD6zpY+/vc2if6rCf/vHFY77j7HbmDUQDt/LzCjypP248Ub62bgHt+REsApXUCrgzIKqgETLturad3B+PvX61za8N951v7w3xrvb0DthbIfNuHXAJu7BFf+0ATMAHbHhG1TyT30/Ggg+lg3Ymhw9rSiw6mg9s6kXxC6HjQwXSw7sTQYW3pRQfTwW2dSD4hdDzoYDpYd2LosLb0ooPp4LZOJJ8QOh50MB2sOzF0WFt60cF0cFsnkk8IHQ86mA7WnRg6rC296OBpd/Dqu0+Rw68GhYNXXq4f4UXOj//fQ171SGzD8tr60GsDrFs6iOvDcPP+k5mnrzYOKZYklq/1xDSxDWHmwcWr84NLz15v3H7+4csch38NEsvENLH9DwLs1co+Fv2iAAAAAElFTkSuQmCC",ue=""+new URL("Retry-8b2ec896.png",import.meta.url).href,pe=""+new URL("Timeout-dcabf0f7.jpg",import.meta.url).href,de="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAABYklEQVRoBe2av0oDQRDGZxbRxhfwDRI0NhKtRAhWPkM6Ex9KTOczWElArBRsAuEeIS+QRpvJfJdcqkWRLWYH5or7s7N797v59j4Odph2m4hw//xywsT3JHQqJMddrIajcq2Jaalcs2bx+cTMAi7Grn9xfSI/388kMsJ19RvznA+Pxs3X+yoh867gkV1NNJjBzr3BcKpT5rH6rOcAmR5SO+dzQQdtYE/4YB2w5hGVPdXmNnnSfCvYUz7kpzVewFor9woc/DeDb/OXX4fcjO728b/67jsWnLhXgHtnw/anqCAJpkPdKxAvYDp/9OHhQtYKhAtZKxAuZK1AuJC1AuFC1gqEC1krEC5krUC4kLUC4ULWCoQLWSsQLmStQLhQKFCYAaxSrgvvYTYc7AnL92YEpQ9WdqxSzkrvYzUe7Lwt8rh6dVMn0WVL6yWaxcdtQtUHCidIG7pY9cddsUfL3sF6LbfZAN5wf/+tIkpkAAAAAElFTkSuQmCC",ye="data:image/png;base64,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",pt="data:image/png;base64,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",me="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAABGdBTUEAALGPC/xhBQAAADhlWElmTU0AKgAAAAgAAYdpAAQAAAABAAAAGgAAAAAAAqACAAQAAAABAAAAMKADAAQAAAABAAAAMAAAAAD4/042AAAEs0lEQVRoBe1azWtUVxQ/781HkslHzQwpDfkQUUpdaHZtaUtTuhACYtC/wI22FHd+bMSlFNSdIhjcddFNKW0pZlfS0BayEdSFqAjRJMbGfBgTZ+JM8p7nd27ezMvkvsy7yUucAS+898479+P8zj3n3nvembGIy8xMttOhwiVy6RuX3HbwqrVYZE2SRUM2Jc5lMqlxaxX8Hdd109UKWofLsqxZVqLHxszXGngoBMzAbsNtdBrWBI+x29Xu8xtNJLDbGzWohbr3CrxrK8W3A4BtW9SYqqdEQg1fKCzT6+wSOY4bubjIFQD41g+ayLZL3hmrS1KSlZmbX4xciZKUiOYGMw/wfz/M0ldXnsgFGjzURV2sfRceF+3KhwPxCYdDQslhml+ImVq54KKlVwv6v7Pd9GFzXIA/f7Ui/T5qidE/Z3bT1MIyfXn5qfRBhb9/ptGmvv11dOLzFCVi0i3ULe560mVEBr/6lN4igW/+Nr5hRU+u8/TlfdlXqychr9QO8tQUTGcd+ul2TmT98EVK31nDtaamX1aWqukYxGpuaqB69nm4zfk/Xkizi0faqPfjFC29ydPCIoPUlH9H83T61gKlUzbdOt6qaaFnRa6AbhFDtOM4FRfxpzdmxNgj32X0aDXcyHchbJXYbTa1jTIa0502cgXUbLuBrqKZxCJrhffEgP2i2Kac2BYFyoWEfmc0pguyqhRwePusaQu4cW9bDW0z2hYLbDYWcmOmDkTRK6DbRsPGQuJC4SdfWm5bLDQ4sURdv07KBbpSLDT8f55c9oc0hxQmxciFCg7RtUdZ+v1ZnqbfOBIz+WMn0HePdhFCtpMjczSe4w6r9NixdprnffLA4CxzAlwlZlF/d530CXszUuDq4yzdfLqkgh+eKMDwLzuhA+ImAEIc5LCfl3YaNFajtNXZ1N+epFN7w8dBGNNIgd+m8gJgoKeFvs4k0H9daeZEDcrAZ61iBY/GcxfX3T8UPkxAn0rFKBb6ZHhW3OZBbzAI3SIGiDCxUCWwunojCzichPHZXzeefHFtOhbSjrgx00gBHDRhCgK6oLA5TH+TNkYKYAFWWzFSgOSory4VjBRQFgjnRjulptGxJ8FWiA9u7ET4tEy3NssFGrytlO9fLNLMynoXNlOAW1daB942iu/iGKdScIFWuaLNK/FnNk/fTr4kPP3FSIG2es7Gs9P99brgH2MN7eWFBl/lqOv+hFygK8VCawYJeIEFYAm/NYwOsh/ncnR9PldMo3hhgHpCqkWjB7uoPRkX4OMFlRfq5ETP2P4Omswv0557Y3IYKoywiAolpDe/+tNQHi1pm7KpznDcdDHdaBZKnNnVwPGMS78s5mlqhUMGBDUiX7mGCFKkwld+R/PVSwDzrQSf3ZPfMaQKRvCCrBEz+Mm/jaHLumJkgXW9NQwvLwS3OTmByJPjoo409bU0bJgX0gy1htX5RI0F5uFUUmYfVjDaRteMGPCCLDQSuQA81tJRbIVYCHVbKZ7bQAGvRK7AlvJCHirN0z/r/urIXcg/+E7QZWt7J0RGK+O9AtHOp/loHKHwfw9qtAC7zefDUI3i5wOOhmr/zx74ywr+9cE5nZ9rwZ2AEViBGdjfAhPs4mowdpbkAAAAAElFTkSuQmCC",dt=""+new URL("VirtualOutbound-3bb05b70.png",import.meta.url).href,yt={class:"policy-type-tag"},mt=["src"],ht=x({__name:"PolicyTypeTag",props:{policyType:{type:String,required:!0}},setup(r){const n=r,u=se(),_={CircuitBreaker:{iconUrl:le},FaultInjection:{iconUrl:ut},HealthCheck:{iconUrl:Ae},MeshAccessLog:{iconUrl:de},MeshCircuitBreaker:{iconUrl:le},MeshGateway:{iconUrl:null},MeshGatewayRoute:{iconUrl:null},MeshHealthCheck:{iconUrl:Ae},MeshProxyPatch:{iconUrl:ce},MeshRateLimit:{iconUrl:re},MeshRetry:{iconUrl:ue},MeshTimeout:{iconUrl:pe},MeshTrace:{iconUrl:me},MeshTrafficPermission:{iconUrl:ye},ProxyTemplate:{iconUrl:ce},RateLimit:{iconUrl:re},Retry:{iconUrl:ue},Timeout:{iconUrl:pe},TrafficLog:{iconUrl:de},TrafficPermission:{iconUrl:ye},TrafficRoute:{iconUrl:pt},TrafficTrace:{iconUrl:me},VirtualOutbound:{iconUrl:dt}},D=R(()=>{const T=u.state.policyTypes.map(A=>{const b=_[A.name]??{iconUrl:null};return[A.name,b]});return Object.fromEntries(T)}),h=R(()=>D.value[n.policyType]);return(T,A)=>(e(),s("span",yt,[h.value.iconUrl!==null?(e(),s("img",{key:0,class:"policy-type-tag-icon",src:h.value.iconUrl,alt:""},null,8,mt)):(e(),m(f(he),{key:1,icon:"brain",size:"24"})),a(),ae(T.$slots,"default",{},()=>[a(p(n.policyType),1)],!0)]))}});const Be=S(ht,[["__scopeId","data-v-0052ac03"]]),vt={class:"policy-type-heading"},gt={class:"policy-list"},ft={key:0},wt=x({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(r){const n=r,u=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function _({headerKey:D}){return{class:`cell-${D}`}}return(D,h)=>{const T=X("router-link");return e(),m($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,G(n.policyTypeEntries,(A,b)=>(e(),m(Z,{key:b},{"accordion-header":t(()=>[y("h3",vt,[o(Be,{"policy-type":A.type},{default:t(()=>[a(p(A.type)+" ("+p(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",gt,[o(f(ve),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:u,"cell-attrs":_,"disable-pagination":"","is-clickable":""},{sourceTags:t(({rowValue:i})=>[i.length>0?(e(),m(j,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),destinationTags:t(({rowValue:i})=>[i.length>0?(e(),m(j,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),name:t(({rowValue:i})=>[i!==null?(e(),s(v,{key:0},[a(p(i),1)],64)):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",ft,[(e(!0),s(v,null,G(i,(w,Q)=>(e(),s("li",{key:`${b}-${Q}`},[o(T,{to:w.route},{default:t(()=>[a(p(w.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:w})=>[i!==null?(e(),m(De,{key:0,id:`${n.id}-${b}-${w}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Tt=S(wt,[["__scopeId","data-v-71c85650"]]),bt={class:"policy-type-heading"},Dt={class:"policy-list"},Bt={key:1,class:"tag-list-wrapper"},Ct={key:0},kt={key:1},_t={key:0},Ot={key:0},Et=x({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(r){const n=r,u=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function _({headerKey:D}){return{class:`cell-${D}`}}return(D,h)=>{const T=X("router-link");return e(),m($,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(v,null,G(n.ruleEntries,(A,b)=>(e(),m(Z,{key:b},{"accordion-header":t(()=>[y("h3",bt,[o(Be,{"policy-type":A.type},{default:t(()=>[a(p(A.type)+" ("+p(A.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",Dt,[o(f(ve),{class:"policy-type-table",fetcher:()=>({data:A.connections,total:A.connections.length}),headers:u,"cell-attrs":_,"disable-pagination":"","is-clickable":""},{type:t(({rowValue:i})=>[i.sourceTags.length===0&&i.destinationTags.length===0?(e(),s(v,{key:0},[a(`
                —
              `)],64)):(e(),s("div",Bt,[i.sourceTags.length>0?(e(),s("div",Ct,[a(`
                  From

                  `),o(j,{class:"tag-list",tags:i.sourceTags},null,8,["tags"])])):Y("",!0),a(),i.destinationTags.length>0?(e(),s("div",kt,[a(`
                  To

                  `),o(j,{class:"tag-list",tags:i.destinationTags},null,8,["tags"])])):Y("",!0)]))]),addresses:t(({rowValue:i})=>[i.length>0?(e(),s("ul",_t,[(e(!0),s(v,null,G(i,(w,Q)=>(e(),s("li",{key:`${b}-${Q}`},p(w),1))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",Ot,[(e(!0),s(v,null,G(i,(w,Q)=>(e(),s("li",{key:`${b}-${Q}`},[o(T,{to:w.route},{default:t(()=>[a(p(w.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:w})=>[i!==null?(e(),m(De,{key:0,id:`${n.id}-${b}-${w}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(v,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const Pt=S(Et,[["__scopeId","data-v-74be3da4"]]),Gt=y("h2",{class:"visually-hidden"},`
    Policies
  `,-1),Qt={key:0,class:"mt-2"},It=y("h2",{class:"mb-2"},`
      Rules
    `,-1),Ut=x({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(r){const n=r;return(u,_)=>(e(),s(v,null,[Gt,a(),o(Tt,{id:"policies","policy-type-entries":n.policyTypeEntries},null,8,["policy-type-entries"]),a(),r.ruleEntries.length>0?(e(),s("div",Qt,[It,a(),o(Pt,{id:"rules","rule-entries":n.ruleEntries},null,8,["rule-entries"])])):Y("",!0)],64))}}),Lt={key:2,class:"policies-list"},zt={key:3,class:"policies-list"},xt=x({__name:"DataplanePolicies",props:{dataplaneOverview:{type:Object,required:!0}},setup(r){const n=r,u=ge(),_=se(),D=N(null),h=N([]),T=N([]),A=N([]),b=N([]),i=N(!0),w=N(null);Ee(()=>n.dataplaneOverview.name,function(){Q()}),Q();async function Q(){var c,d;w.value=null,i.value=!0,h.value=[],T.value=[],A.value=[],b.value=[];try{if(((d=(c=n.dataplaneOverview.dataplane.networking.gateway)==null?void 0:c.type)==null?void 0:d.toUpperCase())==="BUILTIN")D.value=await u.getMeshGatewayDataplane({mesh:n.dataplaneOverview.mesh,name:n.dataplaneOverview.name}),A.value=J(D.value),b.value=W(D.value.policies);else{const{items:l}=await u.getSidecarDataplanePolicies({mesh:n.dataplaneOverview.mesh,name:n.dataplaneOverview.name});h.value=I(l??[]);const{items:B}=await u.getDataplaneRules({mesh:n.dataplaneOverview.mesh,name:n.dataplaneOverview.name});T.value=L(B??[])}}catch(g){g instanceof Error?w.value=g:console.error(g)}finally{i.value=!1}}function J(c){const d=[],g=c.listeners??[];for(const l of g)for(const B of l.hosts)for(const k of B.routes){const E=[];for(const P of k.destinations){const C=W(P.policies),M={routeName:k.route,route:{name:"policy-detail-view",params:{mesh:c.gateway.mesh,policyPath:"meshgatewayroutes",policy:k.route}},service:P.tags["kuma.io/service"],policies:C};E.push(M)}d.push({protocol:l.protocol,port:l.port,hostName:B.hostName,routeEntries:E})}return d}function W(c){if(c===void 0)return[];const d=[];for(const g of Object.values(c)){const l=_.state.policyTypesByName[g.type];d.push({type:g.type,name:g.name,route:{name:"policy-detail-view",params:{mesh:g.mesh,policyPath:l.path,policy:g.name}}})}return d}function I(c){const d=new Map;for(const l of c){const{type:B,service:k}=l,E=typeof k=="string"&&k!==""?[{label:"kuma.io/service",value:k}]:[],P=B==="inbound"||B==="outbound"?l.name:null;for(const[C,M]of Object.entries(l.matchedPolicies)){d.has(C)||d.set(C,{type:C,connections:[]});const q=d.get(C),K=_.state.policyTypesByName[C];for(const ne of M){const z=U(ne,K,l,E,P);q.connections.push(...z)}}}const g=Array.from(d.values());return g.sort((l,B)=>l.type.localeCompare(B.type)),g}function U(c,d,g,l,B){const k=c.conf&&Object.keys(c.conf).length>0?ie(c.conf):null,P=[{name:c.name,route:{name:"policy-detail-view",params:{mesh:c.mesh,policyPath:d.path,policy:c.name}}}],C=[];if(g.type==="inbound"&&Array.isArray(c.sources))for(const{match:M}of c.sources){const K={sourceTags:[{label:"kuma.io/service",value:M["kuma.io/service"]}],destinationTags:l,name:B,config:k,origins:P};C.push(K)}else{const q={sourceTags:[],destinationTags:l,name:B,config:k,origins:P};C.push(q)}return C}function L(c){const d=new Map;for(const l of c){d.has(l.policyType)||d.set(l.policyType,{type:l.policyType,connections:[]});const B=d.get(l.policyType),k=_.state.policyTypesByName[l.policyType],E=O(l,k);B.connections.push(...E)}const g=Array.from(d.values());return g.sort((l,B)=>l.type.localeCompare(B.type)),g}function O(c,d){const{type:g,service:l,subset:B,conf:k}=c,E=B?Object.entries(B):[];let P,C;g==="ClientSubset"?E.length>0?P=E.map(([z,ee])=>({label:z,value:ee})):P=[{label:"kuma.io/service",value:"*"}]:P=[],g==="DestinationSubset"?E.length>0?C=E.map(([z,ee])=>({label:z,value:ee})):typeof l=="string"&&l!==""?C=[{label:"kuma.io/service",value:l}]:C=[{label:"kuma.io/service",value:"*"}]:g==="ClientSubset"&&typeof l=="string"&&l!==""?C=[{label:"kuma.io/service",value:l}]:C=[];const M=c.addresses??[],q=k&&Object.keys(k).length>0?ie(k):null,K=[];for(const z of c.origins)K.push({name:z.name,route:{name:"policy-detail-view",params:{mesh:z.mesh,policyPath:d.path,policy:z.name}}});return[{type:{sourceTags:P,destinationTags:C},addresses:M,config:q,origins:K}]}return(c,d)=>i.value?(e(),m(fe,{key:0})):w.value!==null?(e(),m(we,{key:1,error:w.value},null,8,["error"])):h.value.length>0?(e(),s("div",Lt,[o(Ut,{"dpp-name":n.dataplaneOverview.name,"policy-type-entries":h.value,"rule-entries":T.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):A.value.length>0&&D.value!==null?(e(),s("div",zt,[o(rt,{"mesh-gateway-dataplane":D.value,"mesh-gateway-listener-entries":A.value,"mesh-gateway-route-policies":b.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),m(Te,{key:4}))}});const Mt=S(xt,[["__scopeId","data-v-2a40d801"]]);const Rt={},Yt={class:"definition-list"};function St(r,n){return e(),s("dl",Yt,[ae(r.$slots,"default",{},void 0,!0)])}const Nt=S(Rt,[["render",St],["__scopeId","data-v-48665ce3"]]),Ht={class:"definition-list-item"},qt={class:"definition-list-item__term"},Kt={class:"definition-list-item__details"},jt=x({__name:"DefinitionListItem",props:{term:{type:String,required:!0}},setup(r){const n=r;return(u,_)=>(e(),s("div",Ht,[y("dt",qt,p(n.term),1),a(),y("dd",Kt,[ae(u.$slots,"default",{},void 0,!0)])]))}});const Ft=S(jt,[["__scopeId","data-v-74f2c619"]]),Jt={class:"stack"},Wt={class:"variable-columns"},Vt={class:"status-with-reason"},Xt=["href"],Zt=x({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(r){const n=r,{t:u,formatIsoDate:_}=be(),D=ge(),h=se(),T=[{hash:"#overview",title:u("data-planes.routes.item.tabs.overview")},{hash:"#insights",title:u("data-planes.routes.item.tabs.insights")},{hash:"#dpp-policies",title:u("data-planes.routes.item.tabs.policies")},{hash:"#xds-configuration",title:u("data-planes.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:u("data-planes.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:u("data-planes.routes.item.tabs.clusters")},{hash:"#mtls",title:u("data-planes.routes.item.tabs.mtls")}],A=R(()=>qe(n.dataplaneOverview.dataplane,n.dataplaneOverview.dataplaneInsight)),b=R(()=>oe(n.dataplaneOverview.dataplane)),i=R(()=>Ke(n.dataplaneOverview.dataplaneInsight)),w=R(()=>je(n.dataplaneOverview,_)),Q=R(()=>{var U;const I=Array.from(((U=n.dataplaneOverview.dataplaneInsight)==null?void 0:U.subscriptions)??[]);return I.reverse(),I}),J=R(()=>{var c;const I=((c=n.dataplaneOverview.dataplaneInsight)==null?void 0:c.subscriptions)??[];if(I.length===0)return[];const U=I[I.length-1];if(!("version"in U)||!U.version)return[];const L=[],O=U.version;if(O.kumaDp&&O.envoy){const d=Fe(O);d.kind!==Je&&d.kind!==We&&L.push(d)}return h.getters["config/getMulticlusterStatus"]&&oe(n.dataplaneOverview.dataplane).find(l=>l.label===Pe)&&typeof O.kumaDp.kumaCpCompatible=="boolean"&&!O.kumaDp.kumaCpCompatible&&L.push({kind:Ve,payload:{kumaDp:O.kumaDp.version}}),L});async function W(I){const{mesh:U,name:L}=n.dataplaneOverview;return await D.getDataplaneFromMesh({mesh:U,name:L},I)}return(I,U)=>{const L=X("RouterLink");return e(),m(Se,{tabs:T},{overview:t(()=>[y("div",Jt,[J.value.length>0?(e(),m(He,{key:0,warnings:J.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):Y("",!0),a(),o(f(H),null,{body:t(()=>[y("div",Wt,[o(V,null,{title:t(()=>[a(p(f(u)("http.api.property.status")),1)]),body:t(()=>[y("div",Vt,[o(Le,{status:A.value.status},null,8,["status"]),a(),A.value.reason.length>0?(e(),m(f(Ge),{key:0,label:A.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[o(f(he),{icon:"info",size:"20","hide-title":""})]),_:1},8,["label"])):Y("",!0)])]),_:1}),a(),o(V,null,{title:t(()=>[a(p(f(u)("http.api.property.name")),1)]),body:t(()=>[o(Ne,{text:n.dataplaneOverview.name},{default:t(()=>[o(L,{to:{name:"data-plane-detail-view",params:{mesh:n.dataplaneOverview.mesh,dataPlane:n.dataplaneOverview.name}}},{default:t(()=>[a(p(n.dataplaneOverview.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),_:1}),a(),o(V,null,{title:t(()=>[a(p(f(u)("http.api.property.tags")),1)]),body:t(()=>[b.value.length>0?(e(),m(j,{key:0,tags:b.value},null,8,["tags"])):(e(),s(v,{key:1},[a(p(f(u)("common.detail.none")),1)],64))]),_:1}),a(),o(V,null,{title:t(()=>[a(p(f(u)("http.api.property.dependencies")),1)]),body:t(()=>[i.value!==null?(e(),m(j,{key:0,tags:i.value},null,8,["tags"])):(e(),s(v,{key:1},[a(p(f(u)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),o(Re,{id:"code-block-data-plane",resource:n.dataplaneOverview,"resource-fetcher":W,"is-searchable":""},null,8,["resource"])])]),insights:t(()=>[o(f(H),null,{body:t(()=>[o(Ye,{"is-empty":Q.value.length===0},{default:t(()=>[o($,{"initially-open":0},{default:t(()=>[(e(!0),s(v,null,G(Q.value,(O,c)=>(e(),m(Z,{key:c},{"accordion-header":t(()=>[o(Ie,{subscription:O},null,8,["subscription"])]),"accordion-content":t(()=>[o(Ue,{subscription:O,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),_:1})]),"dpp-policies":t(()=>[o(f(H),null,{body:t(()=>[o(Mt,{"dataplane-overview":r.dataplaneOverview},null,8,["dataplane-overview"])]),_:1})]),"xds-configuration":t(()=>[o(f(H),null,{body:t(()=>[o(te,{"data-path":"xds",mesh:r.dataplaneOverview.mesh,"dpp-name":r.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),"envoy-stats":t(()=>[o(f(H),null,{body:t(()=>[o(te,{"data-path":"stats",mesh:r.dataplaneOverview.mesh,"dpp-name":r.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),"envoy-clusters":t(()=>[o(f(H),null,{body:t(()=>[o(te,{"data-path":"clusters",mesh:r.dataplaneOverview.mesh,"dpp-name":r.dataplaneOverview.name,"query-key":"envoy-data-data-plane"},null,8,["mesh","dpp-name"])]),_:1})]),mtls:t(()=>[o(f(H),null,{body:t(()=>[w.value===null?(e(),m(f(Qe),{key:0,appearance:"danger"},{alertMessage:t(()=>[a(`
              This data plane proxy does not yet have mTLS configured —
              `),y("a",{href:f(u)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},`
                Learn About Certificates in `+p(f(u)("common.product.name")),9,Xt)]),_:1})):(e(),m(Nt,{key:1},{default:t(()=>[(e(!0),s(v,null,G(w.value,(O,c)=>(e(),m(Ft,{key:c,term:f(u)(`http.api.property.${c}`)},{default:t(()=>[a(p(O),1)]),_:2},1032,["term"]))),128))]),_:1}))]),_:1})]),_:1})}}});const $t=S(Zt,[["__scopeId","data-v-56d4ace4"]]),da=x({__name:"DataPlaneDetailView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(r){const n=r,{t:u}=be();return(_,D)=>(e(),m(Me,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(({route:h})=>[o(ze,{breadcrumbs:[{to:{name:`${n.isGatewayView?"gateways":"data-planes"}-list-view`,params:{mesh:h.params.mesh}},text:f(u)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[y("h2",null,[o(Xe,{title:f(u)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:h.params.dataPlane}),render:!0},null,8,["title"])])]),default:t(()=>[a(),o(xe,{src:`/meshes/${h.params.mesh}/dataplane-overviews/${h.params.dataPlane}`},{default:t(({data:T,isLoading:A,error:b})=>[A?(e(),m(fe,{key:0})):b?(e(),m(we,{key:1,error:b},null,8,["error"])):T===void 0?(e(),m(Te,{key:2})):(e(),m($t,{key:3,"dataplane-overview":T,"data-testid":"detail-view-details"},null,8,["dataplane-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{da as default};
